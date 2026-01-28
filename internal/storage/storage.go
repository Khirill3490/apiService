package storage

import "context"

type UrlStorage interface {
	Save(ctx context.Context, originalUrl string) (id int64, shortUrl string, err error)
	Get(ctx context.Context, shortUrl string) (originalUrl string, err error)
	Delete(ctx context.Context, shortUrl string) error
}

type ErrNotFound struct{}

func (e ErrNotFound) Error() string { return "not found" }
