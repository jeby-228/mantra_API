package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set JWT secret for testing
	if err := os.Setenv("JWT_SECRET", "test_secret_key_for_testing"); err != nil {
		t.Fatalf("Failed to set JWT_SECRET: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("JWT_SECRET"); err != nil {
			t.Errorf("Failed to unset JWT_SECRET: %v", err)
		}
	}()

	t.Run("Missing Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)

		AuthMiddleware()(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "缺少 Authorization header")
		assert.True(t, c.IsAborted())
	})

	t.Run("Invalid Authorization Header Format - No Bearer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		c.Request.Header.Set("Authorization", "InvalidToken")

		AuthMiddleware()(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header 格式錯誤")
		assert.True(t, c.IsAborted())
	})

	t.Run("Invalid Authorization Header Format - Only Bearer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer")

		AuthMiddleware()(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header 格式錯誤")
		assert.True(t, c.IsAborted())
	})

	t.Run("Invalid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid_token")

		AuthMiddleware()(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "無效的 token")
		assert.True(t, c.IsAborted())
	})

	t.Run("Valid Token", func(t *testing.T) {
		// Generate a valid token for testing
		token, err := GenerateToken(1, "test@example.com")
		assert.NoError(t, err)

		// Create a router to test the middleware
		router := gin.New()
		nextCalled := false
		router.Use(AuthMiddleware())
		router.GET("/", func(c *gin.Context) {
			nextCalled = true

			// Verify context values
			userID, exists := c.Get("user_id")
			assert.True(t, exists)
			assert.Equal(t, int64(1), userID)

			email, exists := c.Get("user_email")
			assert.True(t, exists)
			assert.Equal(t, "test@example.com", email)

			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)
	})
}
