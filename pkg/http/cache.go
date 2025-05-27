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
	}
)

func (h handler) getCacheKey(req *http.Request) string {
	cacheKey := sha256.New()

	// Include HTTP method in the hash
	cacheKey.Write([]byte(req.Method))
	cacheKey.Write([]byte(":"))

	// Add the path
	cacheKey.Write([]byte(req.URL.Path))

	// Add query parameters if enabled
	if h.cfg.ShouldHashQuery {
		// Sort query parameters to ensure consistent ordering
		query := req.URL.Query()
		keys := make([]string, 0, len(query))
		for k := range query {
			if _, ok := h.cfg.HashQueryIgnore[k]; !ok {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)

		for _, key := range keys {
			values := query[key]
			sort.Strings(values)
			for _, value := range values {
				cacheKey.Write([]byte("&"))
				cacheKey.Write([]byte(key))
				cacheKey.Write([]byte("="))
				cacheKey.Write([]byte(value))
			}
		}
	}

	// Add headers if configured
	if len(h.cfg.HashHeaders) > 0 {
		// Sort headers to ensure consistent ordering
		sort.Strings(h.cfg.HashHeaders)

		for _, headerName := range h.cfg.HashHeaders {
			values := req.Header.Values(headerName)
			if len(values) > 0 {
				sort.Strings(values)
				for _, value := range values {
					cacheKey.Write([]byte("|"))
					cacheKey.Write([]byte(headerName))
					cacheKey.Write([]byte("="))
					cacheKey.Write([]byte(value))
				}
			}
		}
	}

	return fmt.Sprintf("%x", cacheKey.Sum(nil))
}

func serveResponseFromMemory(res http.ResponseWriter, result model.Response) {
	maps.Copy(res.Header(), result.Header)

	res.WriteHeader(result.Status)
	_, err := res.Write(result.Body)

	if err != nil {
		log.Error().Err(err).Msg("Error occurred serving response from memory")
	}
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Error().Err(err).Msg("downstream server is down")
	http.Error(res, "service unavailable", http.StatusBadGateway)
}

func newCacheHandler(c Cacher, m MetricsCollector, w Worker, cfg config.CacheConfig) handler {
	netTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 30,
		Transport: netTransport,
	}

	return handler{
		cacher:           c,
		worker:           w,
		metricsCollector: m,
		cfg:              cfg,
		httpClient:       httpClient,
	}
}

func (h handler) asyncCacheRevalidate(hashKey string, req *http.Request) func() {
	return func() {
		ctx := context.Background()
		newReq := req.WithContext(ctx)

		newReq.URL.Host = h.cfg.DownstreamHost.Host
		newReq.URL.Scheme = h.cfg.DownstreamHost.Scheme
		newReq.RequestURI = ""

		resp, err := h.httpClient.Do(newReq)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Errored when sending request to the server")
			return
		}
		defer resp.Body.Close()

		if err := h.cacheResponse(ctx, hashKey)(resp); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Errored when caching response")
		}
	}
}

func (h handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	logger := log.With().Str("path", req.URL.Path).Str("method", req.Method).Logger()
	logCtx := logger.WithContext(req.Context())

	proxy := httputil.NewSingleHostReverseProxy(h.cfg.DownstreamHost)
	proxy.Transport = h.httpClient.Transport
	proxy.ErrorHandler = errHandler

	// websockets
	if strings.ToLower(req.Header.Get("connection")) == "upgrade" {
		logger.Info().Msg("will not cache websocket request")
		proxy.ServeHTTP(res, req)
		return
	}

	if strings.ToLower(req.Method) != "get" {
		logger.Debug().Msg("will not cache non-GET request")
		proxy.ServeHTTP(res, req)
		return
	}

	hashKey := h.getCacheKey(req)

	result, err := h.cacher.LookUp(logCtx, hashKey)
	if err != nil {
		logger.Warn().Err(err).Msg("lookup error occurred")
	}

	if result == nil {
		proxy.ModifyResponse = h.cacheResponse(logCtx, hashKey)
		logger.Debug().Msg("will cache response from downstream")
		proxy.ServeHTTP(res, req)
		return
	}

	logger.Info().Msg("serving request from memory")
	h.metricsCollector.CacheHit(req.Method, result.Status)

	if result.IsStale() {
		go h.worker.Start(hashKey, h.asyncCacheRevalidate(hashKey, req))
	}
	serveResponseFromMemory(res, *result)
}

func (h handler) cacheResponse(ctx context.Context, hashKey string) func(*http.Response) error {
	return func(response *http.Response) error {
		logger := log.Ctx(ctx)
		logger.Debug().Msg("got response from downstream service")
		h.metricsCollector.CacheMiss(response.Request.Method, response.StatusCode)

		if response.StatusCode >= http.StatusInternalServerError {
			logger.Warn().Msg("won't cache 5XX downstream responses")
			return nil
		}

		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			logger.Error().Err(readErr).Msg("error occurred reading response body")
			return nil
		}

		// Create a new reader for the response body
		response.Body = io.NopCloser(bytes.NewReader(body))

		// Create a copy of the header to prevent modification of the original
		headerCopy := make(http.Header)
		maps.Copy(headerCopy, response.Header)

		entry := model.Response{
			Body:   body,
			Header: headerCopy,
			Status: response.StatusCode,
		}

		if err := h.cacher.Store(ctx, hashKey, &entry); err != nil {
			logger.Error().Err(err).Msg("error occurred storing response in memory")
			return nil
		}

		return nil
	}
}
