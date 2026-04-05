package jwt

import (
	"errors"
	"time"

	"im-chat/internal/infra/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	secretKey []byte
	expire    int
	refresh   int
)

type Claims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

func InitJWT(cfg *config.JWTConfig) {
	secretKey = []byte(cfg.SecretKey)
	expire = cfg.Expire
	refresh = cfg.Refresh
}

func GenerateToken(uid string) (string, error) {
	claims := Claims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
