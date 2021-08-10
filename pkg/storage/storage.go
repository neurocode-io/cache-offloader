package storage

import (
	"context"
	"net/http"
)

type Repository interface {
	LookUp(context.Context, string) (*http.Response, error)
	Store(context.Context, string, []byte) error
	CheckConnection(context.Context) error
}
