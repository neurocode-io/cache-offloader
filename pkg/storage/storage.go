package storage

import (
	"context"
)

type Repository interface {
	LookUp(context.Context, string) (*Response, error)
	Store(context.Context, string, *Response) error
	CheckConnection(context.Context) error
}
