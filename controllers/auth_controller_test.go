package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"mantra_API/audit"
	"mantra_API/auth"
	"mantra_API/internal/testhelper"
	"mantra_API/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

const (
	// #nosec G101 -- This is a public LINE API path, not credentials.
	lineTokenPath   = "/oauth2/v2.1/token"
	lineProfilePath = "/v2/profile"
)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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

func testCreatorUUID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

func withMockLineAPI(
	t *testing.T,
	responder func(*http.Request) (*http.Response, error),
) {
	t.Helper()

	originalTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "api.line.me" {
			return responder(req)
		}
		return originalTransport.RoundTrip(req)
	})
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
	})
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
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
	_, err = uuid.Parse(resp.User.ID)
	assert.NoError(t, err)

	var member models.Member
	err = testDB.Where("email = ?", "jeby@example.com").First(&member).Error
	assert.NoError(t, err)
	assert.Equal(t, audit.SelfRegistrationCreatorID, member.CreatorId)
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
			CreatorId:    testCreatorUUID(),
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
			CreatorId:    testCreatorUUID(),
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
	assert.Equal(t, seed.ID.String(), resp.User.ID)
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
			CreatorId:    testCreatorUUID(),
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
			CreatorId:    testCreatorUUID(),
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&seed).Error)

	router := gin.New()
	router.POST("/auth/line/unbind", func(c *gin.Context) {
		// Simulate auth middleware putting user_id into context.
		c.Set("user_id", seed.ID)
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

func TestLineLogin_InvalidAuthorizationCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)
	t.Setenv("LINE_CHANNEL_ID", "test-channel-id")
	t.Setenv("LINE_CHANNEL_SECRET", "test-channel-secret")

	withMockLineAPI(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == lineTokenPath {
			return jsonResponse(http.StatusBadRequest, `{"error":"invalid_grant"}`), nil
		}
		t.Fatalf("unexpected LINE request path: %s", req.URL.Path)
		return nil, nil
	})

	router := gin.New()
	router.POST("/auth/line", LineLogin)

	req := makeJSONPostRequest(t, "/auth/line", LineLoginRequest{
		Code:        "bad-code",
		RedirectURI: "http://localhost:5173/auth/callback",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "無效的 LINE 授權碼")
}

func TestLineLogin_CreatesMemberOnFirstLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)
	t.Setenv("LINE_CHANNEL_ID", "test-channel-id")
	t.Setenv("LINE_CHANNEL_SECRET", "test-channel-secret")

	withMockLineAPI(t, func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case lineTokenPath:
			return jsonResponse(
				http.StatusOK,
				`{"access_token":"line-access-token","id_token":"line-id-token"}`,
			), nil
		case lineProfilePath:
			return jsonResponse(
				http.StatusOK,
				`{"userId":"line-u-001","displayName":"Line First User","pictureUrl":"https://example.com/p.png"}`,
			), nil
		default:
			t.Fatalf("unexpected LINE request path: %s", req.URL.Path)
			return nil, nil
		}
	})

	router := gin.New()
	router.POST("/auth/line", LineLogin)

	req := makeJSONPostRequest(t, "/auth/line", LineLoginRequest{
		Code:        "valid-code",
		RedirectURI: "http://localhost:5173/auth/callback",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "Line First User", resp.User.Name)
	assert.Equal(t, "line-u-001@line.user", resp.User.Email)

	var member models.Member
	err = testDB.Where("line_id = ?", "line-u-001").First(&member).Error
	assert.NoError(t, err)
	assert.Equal(t, "Line First User", member.Name)
	assert.Equal(t, "line-u-001@line.user", member.Email)
}

func TestBindLine_ConflictWhenLineIDBoundByAnotherMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)
	t.Setenv("LINE_CHANNEL_ID", "test-channel-id")
	t.Setenv("LINE_CHANNEL_SECRET", "test-channel-secret")

	currentMember := models.Member{
		Name:  "Current User",
		Email: "current@example.com",
		Base: models.Base{
			CreationTime: time.Now(),
			CreatorId:    testCreatorUUID(),
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&currentMember).Error)

	existingBound := models.Member{
		Name:   "Already Bound",
		Email:  "already@example.com",
		LineID: "line-conflict-001",
		Base: models.Base{
			CreationTime: time.Now(),
			CreatorId:    testCreatorUUID(),
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&existingBound).Error)

	withMockLineAPI(t, func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case lineTokenPath:
			return jsonResponse(
				http.StatusOK,
				`{"access_token":"line-access-token","id_token":"line-id-token"}`,
			), nil
		case lineProfilePath:
			return jsonResponse(
				http.StatusOK,
				`{"userId":"line-conflict-001","displayName":"Conflict User","pictureUrl":"https://example.com/p.png"}`,
			), nil
		default:
			t.Fatalf("unexpected LINE request path: %s", req.URL.Path)
			return nil, nil
		}
	})

	router := gin.New()
	router.POST("/auth/line/bind", func(c *gin.Context) {
		c.Set("user_id", currentMember.ID)
		BindLine(c)
	})

	req := makeJSONPostRequest(t, "/auth/line/bind", LineLoginRequest{
		Code:        "valid-code",
		RedirectURI: "http://localhost:5173/auth/callback",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "該 LINE 帳號已被其他會員綁定")

	var refreshedCurrent models.Member
	err := testDB.First(&refreshedCurrent, currentMember.ID).Error
	assert.NoError(t, err)
	assert.Empty(t, refreshedCurrent.LineID)
}

func TestBindLine_IdempotentWhenAlreadyBoundToCurrentMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := newAuthControllerTestDB(t)
	SetDB(testDB)
	t.Setenv("LINE_CHANNEL_ID", "test-channel-id")
	t.Setenv("LINE_CHANNEL_SECRET", "test-channel-secret")

	seed := models.Member{
		Name:   "Bound User",
		Email:  "bound@example.com",
		LineID: "line-idempotent-001",
		Base: models.Base{
			CreationTime: time.Now(),
			CreatorId:    testCreatorUUID(),
			IsDeleted:    false,
		},
	}
	assert.NoError(t, testDB.Create(&seed).Error)

	var profileRequestCount int
	var mu sync.Mutex
	withMockLineAPI(t, func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case lineTokenPath:
			return jsonResponse(
				http.StatusOK,
				`{"access_token":"line-access-token","id_token":"line-id-token"}`,
			), nil
		case lineProfilePath:
			mu.Lock()
			profileRequestCount++
			mu.Unlock()
			return jsonResponse(
				http.StatusOK,
				`{"userId":"line-idempotent-001","displayName":"Bound User","pictureUrl":"https://example.com/p.png"}`,
			), nil
		default:
			t.Fatalf("unexpected LINE request path: %s", req.URL.Path)
			return nil, nil
		}
	})

	router := gin.New()
	router.POST("/auth/line/bind", func(c *gin.Context) {
		c.Set("user_id", seed.ID)
		BindLine(c)
	})

	req := makeJSONPostRequest(t, "/auth/line/bind", LineLoginRequest{
		Code:        "valid-code",
		RedirectURI: "http://localhost:5173/auth/callback",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"email":"bound@example.com"`)

	var updated models.Member
	err := testDB.First(&updated, seed.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "line-idempotent-001", updated.LineID)

	mu.Lock()
	assert.Equal(t, 1, profileRequestCount)
	mu.Unlock()
}
