package storage

import (
	"context"
)

type Repository interface {
	LookUp(context.Context, string) ([]byte, error)
	Store(context.Context, string, []byte) error
	CheckConnection(context.Context) error
}
