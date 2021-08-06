package storage

import "net/http"

type Repository interface {
	LookUp(string) (http.Response, error)
	Store(string, []byte) error
}
