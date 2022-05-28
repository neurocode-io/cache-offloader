package model

type Response struct {
	Header map[string][]string
	Body   []byte
	Status int
}
