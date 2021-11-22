package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/metrics"
	"neurocode.io/cache-offloader/pkg/storage"
	"neurocode.io/cache-offloader/pkg/utils"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

func getCacheKey(req *http.Request) string {
	cacheKey := sha1.New()
	cacheKey.Write([]byte(req.URL.Path))

	cacheConfig := config.New().CacheConfig
	if cacheConfig.HashShouldQuery {
		for key, values := range req.URL.Query() {
			if _, ok := cacheConfig.HashQueryIgnore[key]; ok {
				continue
			}
			for _, value := range values {
				cacheKey.Write([]byte(fmt.Sprintf("&%s=%s", key, value)))
			}
		}
	}

	return string(cacheKey.Sum(nil))
}

func shouldProxyRequest(err error, result *storage.Response, failureModeDeny bool) bool {
	return (err == nil && result == nil) || (err != nil && !failureModeDeny)
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Error("Error Occured", rz.Err(err))
	http.Error(res, "Something bad happened", http.StatusBadGateway)
}

func cacheResponse(ctx context.Context, hashKey string, repo storage.Repository, metrics *metrics.MetricCollector, failureModeDeny bool) func(*http.Response) error {
	return func(response *http.Response) error {
		logger := rz.FromCtx(ctx)

		logger.Debug("Got response from downstream service")
		metrics.DownstreamHit(response.StatusCode, response.Request.Method)

		if response.StatusCode >= 500 {
			logger.Warn("Won't cache 5XX downstream responses")
			return nil
		}

		var body []byte
		var err error
		if response.Header.Get("content-encoding") == "gzip" {
			reader, _ := gzip.NewReader(response.Body)
			body, err = io.ReadAll(reader)
		} else {
			body, err = io.ReadAll(response.Body)
		}

		if err != nil {
			if failureModeDeny {
				return err
			}
			return nil
		}

		header := response.Header
		statusCode := response.StatusCode
		newBody := ioutil.NopCloser(bytes.NewReader(body))

		response.Body = newBody

		if err = repo.Store(ctx, hashKey, &storage.Response{Body: body, Header: header, Status: statusCode}); err != nil {
			if failureModeDeny {
				return err
			}
		}

		return nil
	}
}

func writeErrorResponse(res http.ResponseWriter, status int, message string) {
	log.Error(message)

	res.WriteHeader(status)
	res.Write([]byte(message))
}

func serveResponseFromMemory(res http.ResponseWriter, result *storage.Response) {
	for key, values := range result.Header {
		res.Header()[key] = values
	}

	res.WriteHeader(result.Status)
	res.Write(result.Body)
}

func CacheHandler(repo storage.Repository, downstreamURL *url.URL) http.HandlerFunc {
	metrics := metrics.NewMetricCollector()
	serverConfig := config.New().ServerConfig
	cacheConfig := config.New().CacheConfig

	return func(res http.ResponseWriter, req *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(downstreamURL)

		if utils.VariableMatchesRegexIn(req.URL.Path, cacheConfig.IgnorePaths) {
			log.Info(fmt.Sprintf("%v is a ignore endpoint.", req.URL.Path))
			proxy.ServeHTTP(res, req)
			return
		}

		// websocket
		if strings.ToLower(req.Header.Get("connection")) == "upgrade" {
			log.Info("Websocket request")
			proxy.ServeHTTP(res, req)
			return
		}

		proxy.ErrorHandler = errHandler

		if !(strings.ToLower(req.Method) == "post" || strings.ToLower(req.Method) == "patch" || strings.ToLower(req.Method) == "put") {
			log.Debug(fmt.Sprintf("%v is not a POST, PATCH or PUT method. Wont do anything.", req.Method))
			proxy.ServeHTTP(res, req)
			return
		}

		hashKey := getCacheKey(req)

		logger := log.With(rz.Fields(
			rz.String("path", req.URL.Path),
			rz.String("method", req.Method),
			rz.Bool("failure-mode-deny", serverConfig.FailureModeDeny),
		))

		ctx := logger.ToCtx(req.Context())

		result, err := repo.LookUp(ctx, hashKey)

		if err != nil && serverConfig.FailureModeDeny {
			writeErrorResponse(res, http.StatusBadGateway, fmt.Sprintf("Storage did not respond in time or error occured: %v", err))
			return
		}

		if shouldProxyRequest(err, result, serverConfig.FailureModeDeny) {
			proxy.ModifyResponse = cacheResponse(ctx, hashKey, repo, metrics, serverConfig.FailureModeDeny)
			logger.Debug("response from downstream cached")
			proxy.ServeHTTP(res, req)
			return
		}

		logger.Info("serving request from memory")
		metrics.CacheHit(result.Status, req.Method)
		serveResponseFromMemory(res, result)
	}
}
