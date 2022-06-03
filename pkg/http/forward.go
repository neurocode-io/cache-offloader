package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/skerkour/rz"
	"github.com/skerkour/rz/log"
)

func forwardHandler(url *url.URL) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		logger := log.With(rz.Fields(
			rz.String("path", req.URL.Path),
			rz.String("method", req.Method),
		))
		proxy := httputil.NewSingleHostReverseProxy(url)
		logger.Info("will not cache this request")
		proxy.ServeHTTP(res, req)
	}
}
