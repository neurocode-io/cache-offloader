package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rs/zerolog/log"
)

func forwardHandler(url *url.URL) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		logger := log.With().Str("path", req.URL.Path).Str("method", req.Method).Logger()
		proxy := httputil.NewSingleHostReverseProxy(url)
		logger.Info().Msg("will not cache this request")
		proxy.ServeHTTP(res, req)
	}
}
