package jwt

import (
	"AIWallHub/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义声明
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken生成JWT token
func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "AIWallHub",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}

// ParseToken解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return config.JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
