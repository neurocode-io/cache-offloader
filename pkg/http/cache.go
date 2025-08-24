package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sort"
	"strings"
	"time"

	"maps"

	"github.com/rs/zerolog/log"

	"github.com/neurocode-io/cache-offloader/config"
	"github.com/neurocode-io/cache-offloader/pkg/model"
)

//go:generate mockgen -source=./cache.go -destination=./cache-mock_test.go -package=http
type (
	Worker interface {
		Start(string, func())
	}
	Cacher interface {
		LookUp(context.Context, string) (*model.Response, error)
		Store(context.Context, string, *model.Response) error
	}

	MetricsCollector interface {
		CacheHit(method string, statusCode int)
		CacheMiss(method string, statusCode int)
	}

	handler struct {
		cacher           Cacher
		worker           Worker
		metricsCollector MetricsCollector
		cfg              config.CacheConfig
		httpClient       *http.Client
		proxy            *httputil.ReverseProxy
	}
)

type ctxKey string

const (
	ctxKeyCacheFunc ctxKey = "cacheFunc"
	maxCacheBytes          = int64(10 << 20) // 10MiB
)

var hopByHop = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"TE":                  {},
	"Trailers":            {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func stripHopByHop(h http.Header) {
	if c := h.Values("Connection"); len(c) > 0 {
		for _, v := range c {
			for _, tok := range strings.Split(v, ",") {
				delete(h, http.CanonicalHeaderKey(strings.TrimSpace(tok)))
			}
		}
	}
	for k := range hopByHop {
		delete(h, k)
	}
}

func isWebSocket(r *http.Request) bool {
	connVals := strings.ToLower(strings.Join(r.Header.Values("Connection"), ","))
	if !strings.Contains(connVals, "upgrade") {
		return false
	}
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func cloneHeaders(src http.Header) http.Header {
	dst := make(http.Header, len(src))
	for k, v := range src {
		dst[k] = append([]string(nil), v...)
	}
	return dst
}

func (h handler) getCacheKey(req *http.Request) string {
	// Check for global cache keys first
	for pattern, globalKey := range h.cfg.GlobalCacheKeys {
		if strings.HasPrefix(req.URL.Path, pattern) {
			return globalKey
		}
	}

	hash := sha256.New()
	hash.Write([]byte(req.Method))
	hash.Write([]byte(":"))
	hash.Write([]byte(req.URL.Path))

	if h.cfg.ShouldHashQuery {
		query := req.URL.Query()
		keys := make([]string, 0, len(query))
		for k := range query {
			if _, ok := h.cfg.HashQueryIgnore[strings.ToLower(k)]; ok {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			values := query[key]
			sort.Strings(values)
			for _, value := range values {
				hash.Write([]byte("&"))
				hash.Write([]byte(key))
				hash.Write([]byte("="))
				hash.Write([]byte(value))
			}
		}
	}

	if len(h.cfg.HashHeaders) > 0 {
		sort.Strings(h.cfg.HashHeaders)
		for _, headerName := range h.cfg.HashHeaders {
			values := req.Header.Values(headerName)
			if len(values) > 0 {
				sort.Strings(values)
				for _, value := range values {
					hash.Write([]byte("|"))
					hash.Write([]byte(headerName))
					hash.Write([]byte("="))
					hash.Write([]byte(value))
				}
			}
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func serveResponseFromMemory(w http.ResponseWriter, res model.Response) {
	for k, vv := range res.Header {
		w.Header()[k] = append([]string(nil), vv...)
	}
	stripHopByHop(w.Header())
	w.Header().Set("Content-Length", fmt.Sprint(len(res.Body)))
	w.WriteHeader(res.Status)
	if _, err := w.Write(res.Body); err != nil {
		log.Error().Err(err).Msg("write cached body")
	}
}

func errHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Error().Err(err).Msg("downstream server is down")
	http.Error(w, "service unavailable", http.StatusBadGateway)
}

func newCacheHandler(c Cacher, m MetricsCollector, w Worker, cfg config.CacheConfig) handler {
	netTransport := &http.Transport{
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

	httpClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: netTransport,
	}

	rp := httputil.NewSingleHostReverseProxy(cfg.DownstreamHost)
	rp.Transport = netTransport
	rp.ErrorHandler = errHandler

	origDirector := rp.Director
	rp.Director = func(r *http.Request) {
		origDirector(r)
		r.Host = cfg.DownstreamHost.Host
		stripHopByHop(r.Header)
		if r.Header.Get("Range") != "" {
			r.Header.Del("Range")
		}
	}

	rp.ModifyResponse = func(resp *http.Response) error {
		if fn, ok := resp.Request.Context().Value(ctxKeyCacheFunc).(func(*http.Response) error); ok && fn != nil {
			return fn(resp)
		}
		return nil
	}

	return handler{
		cacher:           c,
		worker:           w,
		metricsCollector: m,
		cfg:              cfg,
		httpClient:       httpClient,
		proxy:            rp,
	}
}

func (h handler) asyncCacheRevalidate(key string, orig *http.Request) func() {
	return func() {
		ctx := context.Background()

		u := *orig.URL
		u.Scheme = h.cfg.DownstreamHost.Scheme
		u.Host = h.cfg.DownstreamHost.Host

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		req.Header = cloneHeaders(orig.Header)
		stripHopByHop(req.Header)
		req.Header.Del("Range")
		log.Ctx(orig.Context()).Debug().Str("cachKey", key).Msg("revalidating cache")

		resp, err := h.httpClient.Do(req)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("revalidate request")
			return
		}
		defer resp.Body.Close()

		if err := h.cacheResponse(ctx, key)(resp); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("revalidate cache store")
		}
	}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := log.With().Fields(map[string]any{
		"request": map[string]any{
			"path":        r.URL.Path,
			"method":      strings.ToLower(r.Method),
			"queryParams": r.URL.Query().Encode(),
		},
	}).Logger()
	ctx := logger.WithContext(r.Context())

	if strings.ToUpper(r.Method) != http.MethodGet || r.Header.Get("Range") != "" {
		logger.Debug().Msg("skipping cache")
		h.proxy.ServeHTTP(w, r)
		return
	}

	if isWebSocket(r) {
		logger.Debug().Msg("skipping cache: websocket")
		h.proxy.ServeHTTP(w, r)
		return
	}

	key := h.getCacheKey(r)
	entry, err := h.cacher.LookUp(ctx, key)
	if err != nil {
		logger.Warn().Err(err).Msg("lookup error")
	}

	if entry == nil {
		cr := h.cacheResponse(ctx, key)
		r2 := r.WithContext(context.WithValue(ctx, ctxKeyCacheFunc, cr))
		h.proxy.ServeHTTP(w, r2)
		return
	}

	h.metricsCollector.CacheHit(r.Method, entry.Status)
	if entry.IsStale() {
		go h.worker.Start(key, h.asyncCacheRevalidate(key, r))
	}
	logger.Debug().Str("cacheKey", key).Msg("serving cached response")
	serveResponseFromMemory(w, *entry)
}

func (h handler) cacheResponse(ctx context.Context, key string) func(*http.Response) error {
	return func(resp *http.Response) error {
		lg := log.Ctx(ctx)
		h.metricsCollector.CacheMiss(resp.Request.Method, resp.StatusCode)

		if resp.StatusCode >= http.StatusInternalServerError {
			return nil
		}
		if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent) {
			lg.Debug().Str("cacheKey", key).Int("status", resp.StatusCode).Msg("skipping cache: not 200 or 204")
			return nil
		}
		// TODO: add cache control back in
		// cc := resp.Header.Get("Cache-Control")
		// if strings.Contains(cc, "no-store") || strings.Contains(cc, "private") {
		// 	return nil
		// }
		if resp.Header.Get("Set-Cookie") != "" {
			return nil
		}

		lg.Debug().Str("cacheKey", key).Msg("caching response")

		lr := &io.LimitedReader{R: resp.Body, N: maxCacheBytes + 1}
		body, err := io.ReadAll(lr)
		if err != nil {
			lg.Error().Err(err).Msg("read body")
			return nil
		}
		if lr.N <= 0 {
			lg.Warn().Msg("skip cache: body too large")
			return nil
		}

		resp.Body = io.NopCloser(bytes.NewReader(body))

		headerCopy := make(http.Header, len(resp.Header))
		maps.Copy(headerCopy, resp.Header)
		stripHopByHop(headerCopy)
		delete(headerCopy, "Set-Cookie")

		entry := model.Response{
			Body:   body,
			Header: headerCopy,
			Status: resp.StatusCode,
		}

		if err := h.cacher.Store(ctx, key, &entry); err != nil {
			lg.Error().Err(err).Msg("store cache")
			return nil
		}
		h.metricsCollector.CacheMiss(resp.Request.Method, resp.StatusCode)
		return nil
	}
}
