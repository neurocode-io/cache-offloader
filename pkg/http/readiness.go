package http

import (
	"fmt"
	"net/http"
)

func ReadinessHandler(res http.ResponseWriter, req *http.Request) {
	// TODO add redis check
	fmt.Fprintf(res, "Alive")
}
