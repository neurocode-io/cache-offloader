package http

import (
	"fmt"
	"net/http"
)

func LivenessHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "Alive")
}
