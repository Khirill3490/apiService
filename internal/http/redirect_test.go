package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-project/internal/storage"

	"github.com/go-chi/chi/v5"
)

type fakeStore struct {
	data map[string]string
	err  error
}

func (f *fakeStore) Save(ctx context.Context, originalUrl string) (id int64, shortUrl string, err error) {
	return 0, "", nil
}

func (f *fakeStore) Get(ctx context.Context, shortUrl string) (originalUrl string, err error) {
	if f.err != nil {
		return "", f.err
	}
	if f.data == nil {
		f.data = map[string]string{}
	}
	if url, ok := f.data[shortUrl]; ok {
		return url, nil
	}
	return "", storage.ErrNotFound{}
}

func (f *fakeStore) Delete(ctx context.Context, shortUrl string) error {
	return nil
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestRedirect_Success(t *testing.T) {
	store := &fakeStore{
		data: map[string]string{
			"abc": "https://example.com/aboba",
		},
	}
	h := New(store, newDiscardLogger())

	r := chi.NewRouter()
	r.Get("/{alias}", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rr.Code)
	}

	loc := rr.Header().Get("Location")
	if loc != "https://example.com/path" {
		t.Fatalf("expected Location %q, got %q", "https://example.com/path", loc)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	store := &fakeStore{
		data: map[string]string{},
	}
	h := New(store, newDiscardLogger())

	r := chi.NewRouter()
	r.Get("/{alias}", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestRedirect_InternalError(t *testing.T) {
	store := &fakeStore{
		err: errors.New("db down"),
	}
	h := New(store, newDiscardLogger())

	r := chi.NewRouter()
	r.Get("/{alias}", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Проверяем, что вернулся JSON с internal error (без привязки к точному формату)
	body := rr.Body.String()
	if !strings.Contains(body, "internal error") {
		t.Fatalf("expected body to contain %q, got %q", "internal error", body)
	}
}
