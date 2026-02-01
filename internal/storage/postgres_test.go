package storage

import (
	"context"
	"errors"
	"testing"
	"time"
)

func ensureSchema(t *testing.T, p *Postgres) {
	t.Helper()

	// SQL из твоей миграции 000001_init.up.sql
	const schema = `
CREATE TABLE IF NOT EXISTS urls (
    id         BIGSERIAL PRIMARY KEY,
    alias      TEXT UNIQUE NOT NULL,
    url        TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_urls_alias ON urls(alias);
`
	if _, err := p.db.Exec(schema); err != nil {
		t.Fatalf("failed to ensure schema: %v", err)
	}
}

func cleanup(t *testing.T, p *Postgres) {
	t.Helper()

	// очищаем таблицу между тестами
	_, err := p.db.Exec(`TRUNCATE TABLE urls RESTART IDENTITY;`)
	if err != nil {
		t.Fatalf("failed to cleanup: %v", err)
	}
}

func newTestPostgres(t *testing.T) *Postgres {
	t.Helper()

	p, err := NewPostgres()
	if err != nil {
		t.Fatalf("NewPostgres() failed: %v", err)
	}

	ensureSchema(t, p)
	cleanup(t, p)

	t.Cleanup(func() {
		_ = p.Close()
	})

	return p
}

func TestPostgres_SaveAndGet(t *testing.T) {
	p := newTestPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	original := "https://example.com/path?q=1"

	id, alias, err := p.Save(ctx, original)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}
	if alias == "" {
		t.Fatalf("expected alias not empty")
	}

	got, err := p.Get(ctx, alias)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != original {
		t.Fatalf("expected %q, got %q", original, got)
	}
}

func TestPostgres_Delete_RemovesRecord(t *testing.T) {
	p := newTestPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	original := "https://example.com"
	_, alias, err := p.Save(ctx, original)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := p.Delete(ctx, alias); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err = p.Get(ctx, alias)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// ВАЖНО: у тебя возвращается ErrNotFound{} (значение), а не *ErrNotFound
	var nf ErrNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestPostgres_Delete_NotFound(t *testing.T) {
	p := newTestPostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.Delete(ctx, "no_such_alias")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// ВАЖНО: у тебя возвращается ErrNotFound{} (значение), а не *ErrNotFound
	var nf ErrNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("expected ErrNotFound, got %T: %v", err, err)
	}
}
