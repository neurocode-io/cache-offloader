// forward.go
package http

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var hopByHopForward = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"TE":                  {},
	"Trailers":            {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func stripHopByHopForward(h http.Header) {
	if c := h.Values("Connection"); len(c) > 0 {
		for _, v := range c {
			for _, tok := range strings.Split(v, ",") {
				delete(h, http.CanonicalHeaderKey(strings.TrimSpace(tok)))
			}
		}
	}
	for k := range hopByHopForward {
		delete(h, k)
	}
}

func newForwardHandler(u *url.URL) http.Handler {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          4096,
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
		ForceAttemptHTTP2:     true,
	}

	rp := httputil.NewSingleHostReverseProxy(u)
	rp.Transport = tr
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().Err(err).Msg("downstream server is down (forward)")
		http.Error(w, "service unavailable", http.StatusBadGateway)
	}

	orig := rp.Director
	rp.Director = func(r *http.Request) {
		orig(r)
		r.Host = u.Host
		stripHopByHopForward(r.Header)
	}

	rp.ModifyResponse = func(resp *http.Response) error {
		stripHopByHopForward(resp.Header)
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("path", r.URL.Path).Str("method", r.Method).Msg("forwarding without cache")
		rp.ServeHTTP(w, r)
	})
}
