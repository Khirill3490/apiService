package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

// GenerateAccessToken делает короткоживущий JWT (access).
// Внутри кладём userID в поле sub (subject) и ставим exp.
func GenerateAccessToken(userID int64, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

// ParseAccessToken проверяет подпись/exp и возвращает userID.
func ParseAccessToken(tokenString string, secret []byte) (int64, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			// Защита от "alg: none" и подмены алгоритма.
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	// Subject должен быть числом userID
	var userID int64
	_, err = fmt.Sscanf(claims.Subject, "%d", &userID)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}

	return userID, nil
}
