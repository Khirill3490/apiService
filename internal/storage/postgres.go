package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sql.DB
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

	return &Postgres{db: db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) Save(ctx context.Context, originalUrl string) (int64, string, error) {
	alias := generateAlias(8)

	var id int64
	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO urls(alias, original)
		 VALUES ($1, $2)
		 RETURNING id`,
		alias,
		originalUrl,
	).Scan(&id)

	if err != nil {
		return 0, "", err
	}

	return id, alias, nil
}
