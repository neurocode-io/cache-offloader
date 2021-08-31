package storage

import (
	"context"
)

type Response struct {
	Header map[string][]string
	Body   []byte
	Status int
}

type Repository interface {
	LookUp(context.Context, string) (*Response, error)
	Store(context.Context, string, *Response) error
	CheckConnection(context.Context) error
}
