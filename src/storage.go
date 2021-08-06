package main

import "net/http"

type Repository interface {
	lookUp(string) (http.Response, error)
	store(string, []byte) error
}
