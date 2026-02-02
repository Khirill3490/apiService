package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const RefreshTokenBytes = 32 // 32 bytes => нормальная энтропия

// GenerateRefreshToken возвращает:
// 1) сам refresh token (то, что отдаём клиенту)
// 2) hash(token) (то, что храним в БД)
func GenerateRefreshToken() (token string, tokenHash string, err error) {
	b := make([]byte, RefreshTokenBytes)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	// token для клиента (чтобы было удобно в HTTP/JSON)
	token = base64.RawURLEncoding.EncodeToString(b)

	// hash в БД: sha256 от token bytes
	sum := sha256.Sum256([]byte(token))
	tokenHash = base64.RawURLEncoding.EncodeToString(sum[:])

	return token, tokenHash, nil
}

// HashRefreshToken нужен для /refresh и /logout:
// клиент прислал token -> мы считаем hash и ищем в БД.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
