package utils

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func ValidateAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&AccessTokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return AccessTokenSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}
