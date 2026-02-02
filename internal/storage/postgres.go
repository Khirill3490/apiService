package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"api-project/internal/storage/sqlcdb"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db      *sql.DB
	queries *sqlcdb.Queries
}

func NewPostgres() (*Postgres, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Postgres{
		db:      db,
		queries: sqlcdb.New(db),
	}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) Save(ctx context.Context, originalUrl string) (int64, string, error) {
	alias := generateAlias(8)

	row, err := p.queries.SaveURL(ctx, sqlcdb.SaveURLParams{
		Alias: alias,
		Url:   originalUrl,
	})
	if err != nil {
		return 0, "", err
	}

	return row.ID, row.Alias, nil
}

func (p *Postgres) Get(ctx context.Context, alias string) (string, error) {
	originallUrl, err := p.queries.GetURLByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound{}
		}
		return "", err
	}

	return originallUrl, nil
}

func (p *Postgres) Delete(ctx context.Context, alias string) error {
	rows, err := p.queries.DeleteURLByAlias(ctx, alias)
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound{}
	}

	return err
}


func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row, err := p.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound{}
		}
		return User{}, err
	}

	return User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
	}, nil
}

func (p *Postgres) InsertRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (int64, error) {
	return p.queries.InsertRefreshToken(ctx, sqlcdb.InsertRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}

func (p *Postgres) GetRefreshTokenValid(ctx context.Context, tokenHash string) (RefreshToken, error) {
	row, err := p.queries.GetRefreshTokenValid(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshToken{}, ErrNotFound{}
		}
		return RefreshToken{}, err
	}

	return RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

func (p *Postgres) RevokeRefreshToken(ctx context.Context, id int64, replacedBy *int64) error {
	var rb sql.NullInt64
	if replacedBy != nil {
		rb = sql.NullInt64{Int64: *replacedBy, Valid: true}
	}

	rows, err := p.queries.RevokeRefreshToken(ctx, sqlcdb.RevokeRefreshTokenParams{
		ID:         id,
		ReplacedBy: rb,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound{}
	}
	return nil
}

func (p *Postgres) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	rows, err := p.queries.RevokeRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound{}
	}
	return nil
}

