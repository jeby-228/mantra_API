package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}
	jwtSecret = []byte(secret)
}

// Claims 定義 JWT 的 claims（user_id 為會員 GUID 字串）
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT token
func GenerateToken(userID uuid.UUID, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // 24小時過期

	claims := &Claims{
		UserID: userID.String(),
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "member-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken 驗證 JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// 驗證簽名算法，防止算法替換攻擊
			// 第一步：確保使用 HMAC 算法（拒絕 RSA, ECDSA, none 等）
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// 第二步：只接受 HS256，拒絕 HS384 和 HS512
			// 這確保了與我們的密鑰生成和期望算法的一致性
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if token == nil || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if _, err := uuid.Parse(claims.UserID); err != nil {
		return nil, errors.New("無效的 user_id")
	}

	return claims, nil
}
