package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/model"
)

//go:generate mockgen -source=./cache.go -destination=./cache-mock_test.go -package=http
type Cacher interface {
	LookUp(context.Context, string) (*model.Response, error)
	Store(context.Context, string, *model.Response) error
}

type MetricsCollector interface {
	CacheHit(method string, statusCode int)
	CacheMiss(method string, statusCode int)
}

type handler struct {
	cacher           Cacher
	metricsCollector MetricsCollector
	downstreamURL    url.URL
}

func handleGzipServeErr(err error) {
	if err != nil {
		log.Error().Err(err).Msg("Error occurred serving gzip response from memory")
	}
}

func getCacheKey(req *http.Request) string {
	cacheKey := sha256.New()
	cacheKey.Write([]byte(req.URL.Path))

	cacheConfig := config.New().CacheConfig

	if !cacheConfig.ShouldHashQuery {
		return fmt.Sprintf("% x", cacheKey.Sum(nil))
	}

	for key, values := range req.URL.Query() {
		if _, ok := cacheConfig.HashQueryIgnore[key]; ok {
			continue
		}
		for _, value := range values {
			cacheKey.Write([]byte(fmt.Sprintf("&%s=%s", key, value)))
		}
	}

	return fmt.Sprintf("% x", cacheKey.Sum(nil))
}

func serveResponseFromMemory(res http.ResponseWriter, result *model.Response) {
	for key, values := range result.Header {
		res.Header()[key] = values
	}

	res.WriteHeader(result.Status)

	if res.Header().Get("content-encoding") == "gzip" {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		_, err := gz.Write(result.Body)
		handleGzipServeErr(err)
		err = gz.Close()
		handleGzipServeErr(err)
		_, err = res.Write(b.Bytes())
		handleGzipServeErr(err)

		return
	}

	_, err := res.Write(result.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error occurred serving response from memory")
	}
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Error().Err(err).Msg("downstream server is down")
	http.Error(res, "service unavailable", http.StatusBadGateway)
}

func newCacheHandler(c Cacher, m MetricsCollector, url url.URL) handler {
	return handler{
		cacher:           c,
		metricsCollector: m,
		downstreamURL:    url,
	}
}

func (h handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(&h.downstreamURL)
	proxy.ErrorHandler = errHandler
	logger := log.With().Str("path", req.URL.Path).Str("method", req.Method).Logger()
	logCtx := logger.WithContext(req.Context())

	// websockets
	if strings.ToLower(req.Header.Get("connection")) == "upgrade" {
		logger.Info().Msg("will not cache websocket request")
		proxy.ServeHTTP(res, req)

		return
	}

	if !(strings.ToLower(req.Method) == "get") {
		logger.Debug().Msg("will not cache non-GET request")
		proxy.ServeHTTP(res, req)

		return
	}

	hashKey := getCacheKey(req)

	result, err := h.cacher.LookUp(logCtx, hashKey)
	if err != nil {
		writeErrorResponse(res, http.StatusBadGateway, fmt.Errorf("storage error occurred: %v", err))

		return
	}

	if result == nil {
		proxy.ModifyResponse = h.cacheResponse(logCtx, hashKey)
		logger.Debug().Msg("will cache response from downstream")
		proxy.ServeHTTP(res, req)

		return
	}

	logger.Info().Msg("serving request from memory")
	h.metricsCollector.CacheHit(req.Method, result.Status)

	// TODO
	// if result.IsStale() => async fetch from downstream
	serveResponseFromMemory(res, result)
}

func (h handler) cacheResponse(ctx context.Context, hashKey string) func(*http.Response) error {
	return func(response *http.Response) error {
		// if this function returns an error, the proxy will return a 502 Bad Gateway error to the client
		// please see the proxy.ModifyResponse documentation for more information
		logger := log.Ctx(ctx)

		logger.Debug().Msg("Got response from downstream service")
		h.metricsCollector.CacheMiss(response.Request.Method, response.StatusCode)

		if response.StatusCode >= http.StatusInternalServerError {
			logger.Warn().Msg("Won't cache 5XX downstream responses")

			return nil
		}

		var body []byte
		var readErr error
		if response.Header.Get("content-encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body)
			if err != nil {
				logger.Error().Err(err).Msg("Error occurred creating gzip reader")

				return nil
			}
			body, readErr = io.ReadAll(reader)
		} else {
			body, readErr = io.ReadAll(response.Body)
		}

		if readErr != nil {
			logger.Error().Err(readErr).Msg("Error occurred reading response body")

			return nil
		}

		header := response.Header
		statusCode := response.StatusCode
		newBody := ioutil.NopCloser(bytes.NewReader(body))

		response.Body = newBody

		entry := model.Response{Body: body, Header: header, Status: statusCode}

		if err := h.cacher.Store(ctx, hashKey, &entry); err != nil {
			logger.Error().Err(err).Msg("Error occurred storing response in memory")

			return nil
		}

		return nil
	}
}

func writeErrorResponse(res http.ResponseWriter, status int, err error) {
	log.Error().Err(err).Msg("error occurred")

	res.WriteHeader(status)
	_, errWrite := res.Write([]byte("error occurred"))
	if errWrite != nil {
		log.Error().Err(errWrite).Msg("Error occurred writing error response")
	}
}
