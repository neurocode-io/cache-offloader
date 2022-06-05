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

	"github.com/skerkour/rz"
	"github.com/skerkour/rz/log"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/model"
)

//go:generate mockgen -source=./stale-while-revalidate.go -destination=./mock_test.go -package=http
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
	MetricsCollector MetricsCollector
	downstreamURL    url.URL
}

func handleGzipServeErr(err error) {
	if err != nil {
		log.Error("Error occurred serving gzip response from memory", rz.Err(err))
	}
}

func getCacheKey(req *http.Request) string {
	cacheKey := sha256.New()
	cacheKey.Write([]byte(req.URL.Path))

	cacheConfig := config.New().CacheConfig

	if !cacheConfig.HashShouldQuery {
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
		log.Error("Error occurred serving response from memory", rz.Err(err))
	}
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Error("Error occurred", rz.Err(err))
	http.Error(res, "Something bad happened", http.StatusBadGateway)
}

func newStaleWhileRevalidateHandler(c Cacher, m MetricsCollector, url url.URL) handler {
	return handler{
		cacher:           c,
		MetricsCollector: m,
		downstreamURL:    url,
	}
}

func (h handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(&h.downstreamURL)
	proxy.ErrorHandler = errHandler
	logger := log.With(rz.Fields(
		rz.String("path", req.URL.Path),
		rz.String("method", req.Method),
	))
	ctx := logger.ToCtx(req.Context())

	// websockets
	if strings.ToLower(req.Header.Get("connection")) == "upgrade" {
		logger.Info("will not cache websocket request")
		proxy.ServeHTTP(res, req)

		return
	}

	if !(strings.ToLower(req.Method) == "get") {
		logger.Debug("will not cache non-GET request")
		proxy.ServeHTTP(res, req)

		return
	}

	hashKey := getCacheKey(req)

	result, err := h.cacher.LookUp(ctx, hashKey)
	if err != nil {
		writeErrorResponse(res, http.StatusBadGateway, fmt.Sprintf("Storage did not respond in time or error occurred: %v", err))

		return
	}

	if result == nil {
		proxy.ModifyResponse = h.cacheResponse(ctx, hashKey)
		logger.Debug("response from downstream cached")
		proxy.ServeHTTP(res, req)

		return
	}

	logger.Info("serving request from memory")
	h.MetricsCollector.CacheHit(req.Method, result.Status)
	serveResponseFromMemory(res, result)
}

func (h handler) cacheResponse(ctx context.Context, hashKey string) func(*http.Response) error {
	return func(response *http.Response) error {
		logger := rz.FromCtx(ctx)

		logger.Debug("Got response from downstream service")
		h.MetricsCollector.CacheMiss(response.Request.Method, response.StatusCode)

		if response.StatusCode >= http.StatusInternalServerError {
			logger.Warn("Won't cache 5XX downstream responses")

			return nil
		}

		var body []byte
		var readErr error
		if response.Header.Get("content-encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body)
			if err != nil {
				return err
			}
			body, readErr = io.ReadAll(reader)
		} else {
			body, readErr = io.ReadAll(response.Body)
		}

		if readErr != nil {
			return readErr
		}

		header := response.Header
		statusCode := response.StatusCode
		newBody := ioutil.NopCloser(bytes.NewReader(body))

		response.Body = newBody

		if err := h.cacher.Store(ctx, hashKey, &model.Response{Body: body, Header: header, Status: statusCode}); err != nil {
			return err
		}

		return nil
	}
}

func writeErrorResponse(res http.ResponseWriter, status int, message string) {
	log.Error(message)

	res.WriteHeader(status)
	_, err := res.Write([]byte(message))
	if err != nil {
		log.Error("Error occurred writing error response", rz.Err(err))
	}
}
