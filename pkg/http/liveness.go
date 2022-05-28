package http

import (
	"fmt"
	"net/http"
)

func livenessHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "Alive")
}
