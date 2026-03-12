package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"

	"mantra_API/auth"
	"mantra_API/models"
	"mantra_API/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type RegisterRequest struct {
	Name     string `json:"name"     binding:"required"       example:"張三"`
	Email    string `json:"email"    binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type LineLoginRequest struct {
	Code        string `json:"code"         binding:"required" example:"auth_code_from_line"`
	RedirectURI string `json:"redirect_uri" binding:"required" example:"http://localhost:5173/auth/callback"`
}

type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  User   `json:"user"`
}

// Register 用戶註冊
// @Summary 用戶註冊
// @Description 註冊新用戶，返回 JWT token 和用戶信息
// @Tags 認證
// @Accept json
// @Produce json
// @Param register body RegisterRequest true "註冊信息"
// @Success 201 {object} AuthResponse "註冊成功"
// @Failure 400 {object} map[string]string "請求參數錯誤"
// @Failure 409 {object} map[string]string "該電子郵件已被註冊"
// @Failure 500 {object} map[string]string "服務器錯誤"
// @Router /register [post]
func Register(input *gin.Context) {
	if db == nil {
		input.JSON(http.StatusInternalServerError, gin.H{"error": "數據庫連接未配置"})
		return
	}

	var req RegisterRequest
	if err := input.ShouldBindJSON(&req); err != nil {
		input.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用 Service 層建立會員（自動處理密碼加密、審計欄位等）
	svc := services.NewMemberService(db)

	// 註冊時使用 creatorId = 0 表示自行註冊
	member, err := svc.CreateMember(req.Name, req.Email, req.Password, 0)
	if err != nil {
		if err.Error() == "email 已被使用" {
			input.JSON(http.StatusConflict, gin.H{"error": "該電子郵件已被註冊"})
			return
		}
		input.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// #nosec G115 - member.ID 來自資料庫，不會溢位
	user := User{ID: int64(member.ID), Name: member.Name, Email: member.Email}

	// 生成 token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		input.JSON(http.StatusInternalServerError, gin.H{"error": "Token 生成失敗"})
		return
	}

	input.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login 用戶登入
// @Summary 用戶登入
// @Description 用戶登入，驗證郵件和密碼後返回 JWT token 和用戶信息
// @Tags 認證
// @Accept json
// @Produce json
// @Param login body LoginRequest true "登入信息"
// @Success 200 {object} AuthResponse "登入成功"
// @Failure 400 {object} map[string]string "請求參數錯誤"
// @Failure 401 {object} map[string]string "電子郵件或密碼錯誤"
// @Failure 500 {object} map[string]string "服務器錯誤"
// @Router /login [post]
func Login(input *gin.Context) {
	if db == nil {
		input.JSON(http.StatusInternalServerError, gin.H{"error": "數據庫連接未配置"})
		return
	}

	var req LoginRequest
	if err := input.ShouldBindJSON(&req); err != nil {
		input.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查詢用戶
	var member models.Member
	err := db.WithContext(input.Request.Context()).
		Where("email = ?", req.Email).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			input.JSON(http.StatusUnauthorized, gin.H{"error": "電子郵件或密碼錯誤"})
			return
		}
		input.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 驗證密碼
	if !auth.CheckPassword(req.Password, member.PasswordHash) {
		input.JSON(http.StatusUnauthorized, gin.H{"error": "電子郵件或密碼錯誤"})
		return
	}

	// #nosec G115 - member.ID 來自資料庫，不會溢位
	user := User{ID: int64(member.ID), Name: member.Name, Email: member.Email}

	// 生成 token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		input.JSON(http.StatusInternalServerError, gin.H{"error": "Token 生成失敗"})
		return
	}

	input.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile 獲取當前用戶信息（需要認證）
// @Summary 獲取當前用戶信息
// @Description 獲取當前登入用戶的詳細信息，需要 JWT 認證
// @Tags 用戶
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]User "獲取成功"
// @Failure 401 {object} map[string]string "未認證"
// @Failure 404 {object} map[string]string "用戶不存在"
// @Failure 500 {object} map[string]string "服務器錯誤"
// @Router /profile [get]
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認證"})
		return
	}

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "數據庫連接未配置"})
		return
	}

	idValue, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認證"})
		return
	}

	var member models.Member
	if err := db.WithContext(c.Request.Context()).
		Select("id", "name", "email").
		First(&member, idValue).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用戶不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// #nosec G115 - member.ID 來自資料庫，不會溢位
	c.JSON(
		http.StatusOK,
		gin.H{"user": User{ID: int64(member.ID), Name: member.Name, Email: member.Email}},
	)
}

// LineLogin 處理 LINE 登入
// @Summary LINE 登入
// @Description 使用 LINE 授權碼登入，返回 JWT token 和用戶信息
// @Tags 認證
// @Accept json
// @Produce json
// @Param lineLogin body LineLoginRequest true "LINE 登入信息"
// @Success 200 {object} AuthResponse "登入成功"
// @Failure 400 {object} map[string]string "請求參數錯誤"
// @Failure 401 {object} map[string]string "無效的 LINE 授權碼"
// @Failure 500 {object} map[string]string "服務器錯誤"
// @Router /auth/line [post]
func LineLogin(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "數據庫連接未配置"})
		return
	}

	var req LineLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Exchange Code for Token
	channelID := os.Getenv("LINE_CHANNEL_ID")
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")

	if channelID == "" || channelSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LINE 配置未設定"})
		return
	}

	tokenURL := "https://api.line.me/oauth2/v2.1/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", req.Code)
	data.Set("redirect_uri", req.RedirectURI)
	data.Set("client_id", channelID)
	data.Set("client_secret", channelSecret)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法連接到 LINE"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的 LINE 授權碼"})
		return
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法解析 LINE 回應"})
		return
	}

	// 2. Get User Profile from LINE
	profileURL := "https://api.line.me/v2/profile"
	reqProfile, _ := http.NewRequest("GET", profileURL, nil)
	reqProfile.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	client := &http.Client{}
	respProfile, err := client.Do(reqProfile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取 LINE 個人資料"})
		return
	}
	defer respProfile.Body.Close()

	var lineProfile struct {
		UserID      string `json:"userId"`
		DisplayName string `json:"displayName"`
		PictureURL  string `json:"pictureUrl"`
	}
	if err := json.NewDecoder(respProfile.Body).Decode(&lineProfile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法解析 LINE 個人資料"})
		return
	}

	// 3. Find or Create User
	var member models.Member
	result := db.WithContext(c.Request.Context()).Where("line_id = ?", lineProfile.UserID).First(&member)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new user
		// Note: LINE doesn't always provide email. Using placeholder email.
		member = models.Member{
			Name:   lineProfile.DisplayName,
			Email:  lineProfile.UserID + "@line.user",
			LineID: lineProfile.UserID,
		}
		if err := db.WithContext(c.Request.Context()).Create(&member).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法創建用戶"})
			return
		}
	}

	// 4. Generate App Token
	// #nosec G115 - member.ID 來自資料庫，不會溢位
	user := User{ID: int64(member.ID), Name: member.Name, Email: member.Email}
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token 生成失敗"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}
