package storage

import (
	"context"
)

type Repository interface {
	LookUp(ctx context.Context, key string) (*Response, error)
	Store(ctx context.Context, key string, resp *Response) error
	CheckConnection(ctx context.Context) error
}
