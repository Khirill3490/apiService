package storage

import (
	"context"
	"time"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	ExpiresAt time.Time
}

type AuthStorage interface {
	GetUserByUsername(ctx context.Context, username string) (User, error)

	InsertRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (int64, error)
	GetRefreshTokenValid(ctx context.Context, tokenHash string) (RefreshToken, error)

	RevokeRefreshToken(ctx context.Context, id int64, replacedBy *int64) error
	RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error
}

// Общий интерфейс: и urls, и auth (чтобы handler принимал один store)
type Store interface {
	UrlStorage
	AuthStorage
}
