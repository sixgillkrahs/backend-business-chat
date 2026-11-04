package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Ip        string `json:"ip"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int64, username string, role string, ip string, secretKey string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		Ip:        ip,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func GenerateRefreshToken(userID int64, username string, role string, ip string, secretKey string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		Ip:        ip,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func GenerateToken(userID int64, username string, role string, ip string, secretKey string, duration time.Duration) (string, error) {
	return GenerateAccessToken(userID, username, role, ip, secretKey, duration)
}

func ValidateToken(tokenStr string, secretKey string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("phương thức ký token không hợp lệ")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token không hợp lệ hoặc đã hết hạn")
	}

	return claims, nil
}

func ValidateRefreshToken(tokenStr string, secretKey string) (*CustomClaims, error) {
	claims, err := ValidateToken(tokenStr, secretKey)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, errors.New("token không phải là refresh token hợp lệ")
	}
	return claims, nil
}
