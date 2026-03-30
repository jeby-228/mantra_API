package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"mantra_API/auth"
	"mantra_API/internal/testhelper"
	"mantra_API/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type failingCloseReadCloser struct {
	*bytes.Reader
	closeErr error
}

func (r *failingCloseReadCloser) Close() error {
	return r.closeErr
}

func newAuthControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	return testhelper.NewSQLiteTestDB(t)
}

func makeJSONPostRequest(t *testing.T, path string, body any) *http.Request {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body failed: %v", err)
	}

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		path,
		bytes.NewBuffer(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func uintIDToInt64(t *testing.T, v uint) int64 {
	t.Helper()

	id, err := strconv.ParseInt(strconv.FormatUint(uint64(v), 10), 10, 64)
	if err != nil {
		t.Fatalf("convert uint id to int64 failed: %v", err)
	}

	return id
}

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)

	router := gin.New()
	router.POST("/register", Register)

	req := makeJSONPostRequest(t, "/register", RegisterRequest{
		Name:     "Jeby",
		Email:    "jeby@example.com",
		Password: "password123",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "Jeby", resp.User.Name)
	assert.Equal(t, "jeby@example.com", resp.User.Email)

	var member models.Member
	err = testDB.Where("email = ?", "jeby@example.com").First(&member).Error
	assert.NoError(t, err)
	assert.NotEqual(t, "password123", member.PasswordHash)
	assert.NotEmpty(t, member.PasswordHash)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)

	hash, err := auth.HashPassword("password123")
	assert.NoError(t, err)

	seed := models.Member{
		Name:         "Existing",
		Email:        "exists@example.com",
		PasswordHash: hash,
		Base: models.Base{
			CreationTime: time.Now(),
			CreatorId:    1,
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&seed).Error)

	router := gin.New()
	router.POST("/register", Register)

	req := makeJSONPostRequest(t, "/register", RegisterRequest{
		Name:     "Another",
		Email:    "exists@example.com",
		Password: "password123",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "該電子郵件已被註冊")
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)

	hash, err := auth.HashPassword("password123")
	assert.NoError(t, err)

	seed := models.Member{
		Name:         "Tester",
		Email:        "tester@example.com",
		PasswordHash: hash,
		Base: models.Base{
			CreationTime: time.Now(),
			CreatorId:    1,
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&seed).Error)

	router := gin.New()
	router.POST("/login", Login)

	req := makeJSONPostRequest(t, "/login", LoginRequest{
		Email:    "tester@example.com",
		Password: "password123",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "Tester", resp.User.Name)
	assert.Equal(t, "tester@example.com", resp.User.Email)
}

func TestLogin_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)

	hash, err := auth.HashPassword("password123")
	assert.NoError(t, err)

	seed := models.Member{
		Name:         "Tester",
		Email:        "tester2@example.com",
		PasswordHash: hash,
		Base: models.Base{
			CreationTime: time.Now(),
			CreatorId:    1,
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&seed).Error)

	router := gin.New()
	router.POST("/login", Login)

	req := makeJSONPostRequest(t, "/login", LoginRequest{
		Email:    "tester2@example.com",
		Password: "wrong-password",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "電子郵件或密碼錯誤")
}

func TestUnbindLine_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)

	seed := models.Member{
		Name:   "LineUser",
		Email:  "line-user@example.com",
		LineID: "line-abc-123",
		Base: models.Base{
			CreationTime: time.Now(),
			CreatorId:    1,
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&seed).Error)

	router := gin.New()
	router.POST("/auth/line/unbind", func(c *gin.Context) {
		// Simulate auth middleware putting user_id into context.
		c.Set("user_id", uintIDToInt64(t, seed.ID))
		UnbindLine(c)
	})

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/auth/line/unbind",
		http.NoBody,
	)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Member
	err := testDB.First(&updated, seed.ID).Error
	assert.NoError(t, err)
	assert.Empty(t, updated.LineID)
}

func TestBindLine_UnauthorizedWhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)

	router := gin.New()
	router.POST("/auth/line/bind", BindLine)

	req := makeJSONPostRequest(t, "/auth/line/bind", LineLoginRequest{
		Code:        "fake-code",
		RedirectURI: "http://localhost:5173/auth/callback",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "未認證")
}

func TestGetLineProfile_LogsCloseErrorsAndReturnsProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("LINE_CHANNEL_ID", "test-channel-id")
	t.Setenv("LINE_CHANNEL_SECRET", "test-channel-secret")

	oldTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = oldTransport
	})

	const tokenBody = `{"access_token":"access-token","id_token":"id-token"}`
	const profileBody = `{"userId":"line-user-1","displayName":"LINE User","pictureUrl":"https://example.com/avatar.png"}`

	callCount := 0
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		callCount++

		switch callCount {
		case 1:
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "https://api.line.me/oauth2/v2.1/token", req.URL.String())
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body: &failingCloseReadCloser{
					Reader:   bytes.NewReader([]byte(tokenBody)),
					closeErr: errors.New("token-close-error"),
				},
			}, nil
		case 2:
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "https://api.line.me/v2/profile", req.URL.String())
			assert.Equal(t, "Bearer access-token", req.Header.Get("Authorization"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body: &failingCloseReadCloser{
					Reader:   bytes.NewReader([]byte(profileBody)),
					closeErr: errors.New("profile-close-error"),
				},
			}, nil
		default:
			t.Fatalf("unexpected outbound call count: %d", callCount)
			return nil, nil
		}
	})

	previousLogWriter := log.Writer()
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	t.Cleanup(func() {
		log.SetOutput(previousLogWriter)
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/line", http.NoBody)

	profile, err := getLineProfile(c, "valid-code", "http://localhost:5173/auth/callback")
	assert.NoError(t, err)
	if assert.NotNil(t, profile) {
		assert.Equal(t, "line-user-1", profile.UserID)
		assert.Equal(t, "LINE User", profile.DisplayName)
		assert.Equal(t, "https://example.com/avatar.png", profile.PictureURL)
	}

	assert.Equal(t, 2, callCount)
	assert.Contains(t, logBuffer.String(), "close LINE token response body: token-close-error")
	assert.Contains(t, logBuffer.String(), "close LINE profile response body: profile-close-error")
}
