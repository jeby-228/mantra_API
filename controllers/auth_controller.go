package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

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

	// 1. Get LINE Profile
	lineProfile, err := getLineProfile(c, req.Code, req.RedirectURI)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "無效的 LINE 授權碼" || err.Error() == "LINE 配置未設定" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	// 2. Find or Create User
	var member models.Member
	result := db.WithContext(c.Request.Context()).
		Where("line_id = ?", lineProfile.UserID).
		First(&member)

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

// LineProfile 定義 LINE 個人資料結構
type LineProfile struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	PictureURL  string `json:"pictureUrl"`
}

// getLineProfile 處理 LINE 授權碼交換與獲取個人資料的邏輯
func getLineProfile(c *gin.Context, code, redirectURI string) (*LineProfile, error) {
	// 1. Exchange Code for Token
	channelID := os.Getenv("LINE_CHANNEL_ID")
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")

	if channelID == "" || channelSecret == "" {
		return nil, errors.New("LINE 配置未設定")
	}

	tokenURL := "https://api.line.me/oauth2/v2.1/token" //nolint:gosec
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", channelID)
	data.Set("client_secret", channelSecret)

	reqToken, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodPost,
		tokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, errors.New("無法建立請求")
	}
	reqToken.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(reqToken)
	if err != nil {
		return nil, errors.New("無法連接到 LINE")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("無效的 LINE 授權碼")
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, errors.New("無法解析 LINE 回應")
	}

	// 2. Get User Profile from LINE
	profileURL := "https://api.line.me/v2/profile"
	reqProfile, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodGet,
		profileURL,
		nil,
	)
	if err != nil {
		return nil, errors.New("無法建立請求")
	}
	reqProfile.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	client := &http.Client{}
	respProfile, err := client.Do(reqProfile)
	if err != nil {
		return nil, errors.New("無法獲取 LINE 個人資料")
	}
	defer respProfile.Body.Close()

	var lineProfile LineProfile
	if err := json.NewDecoder(respProfile.Body).Decode(&lineProfile); err != nil {
		return nil, errors.New("無法解析 LINE 個人資料")
	}

	return &lineProfile, nil
}

// BindLine 綁定 LINE 帳號
// @Summary 綁定 LINE 帳號
// @Description 將當前登入用戶綁定到 LINE 帳號
// @Tags 認證
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param lineBind body LineLoginRequest true "LINE 綁定信息"
// @Success 200 {object} map[string]User "綁定成功"
// @Failure 400 {object} map[string]string "請求參數錯誤"
// @Failure 401 {object} map[string]string "未認證或 LINE 授權失敗"
// @Failure 409 {object} map[string]string "該 LINE 帳號已被其他會員綁定"
// @Failure 500 {object} map[string]string "服務器錯誤"
// @Router /auth/line/bind [post]
func BindLine(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認證"})
		return
	}
	idValue, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認證"})
		return
	}

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "數據庫連接未配置"})
		return
	}

	var req LineLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lineProfile, err := getLineProfile(c, req.Code, req.RedirectURI)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "無效的 LINE 授權碼" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	// 檢查 LINE ID 是否已被其他用戶綁定
	var existingMember models.Member
	err = db.WithContext(c.Request.Context()).
		Where("line_id = ?", lineProfile.UserID).
		First(&existingMember).Error

	if err == nil {
		// 如果找到記錄，且不是當前用戶，則報錯
		// #nosec G115 - ID 來自資料庫，不會溢位
		if int64(existingMember.ID) != idValue {
			c.JSON(http.StatusConflict, gin.H{"error": "該 LINE 帳號已被其他會員綁定"})
			return
		}
		// 如果是當前用戶，視為成功（idempotent）
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 綁定 LINE ID 到當前用戶
	var member models.Member
	if err := db.WithContext(c.Request.Context()).First(&member, idValue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用戶不存在"})
		return
	}

	member.LineID = lineProfile.UserID
	if err := db.WithContext(c.Request.Context()).Save(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法更新用戶資料"})
		return
	}

	// #nosec G115 - member.ID 來自資料庫，不會溢位
	c.JSON(http.StatusOK, gin.H{
		"user": User{ID: int64(member.ID), Name: member.Name, Email: member.Email},
	})
}

// UnbindLine 解除綁定 LINE 帳號
// @Summary 解除綁定 LINE 帳號
// @Description 解除當前登入用戶的 LINE 帳號綁定
// @Tags 認證
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]User "解除綁定成功"
// @Failure 401 {object} map[string]string "未認證"
// @Failure 500 {object} map[string]string "服務器錯誤"
// @Router /auth/line/unbind [post]
func UnbindLine(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認證"})
		return
	}
	idValue, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未認證"})
		return
	}

	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "數據庫連接未配置"})
		return
	}

	var member models.Member
	if err := db.WithContext(c.Request.Context()).First(&member, idValue).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用戶不存在"})
		return
	}

	// 清除 LineID，使用 NULL 避免 unique index 在空字串上衝突。
	if err := db.WithContext(c.Request.Context()).Model(&member).Update("line_id", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法更新用戶資料"})
		return
	}

	member.LineID = ""

	// #nosec G115 - member.ID 來自資料庫，不會溢位
	c.JSON(http.StatusOK, gin.H{
		"user": User{ID: int64(member.ID), Name: member.Name, Email: member.Email},
	})
}
