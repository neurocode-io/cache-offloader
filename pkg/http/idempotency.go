package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"context"

	"dpd.de/idempotency-offloader/config"
	"dpd.de/idempotency-offloader/pkg/metrics"
	"dpd.de/idempotency-offloader/pkg/storage"
	"dpd.de/idempotency-offloader/pkg/utils"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

func shouldProxyRequest(err error, result *storage.Response, failureModeDeny bool) bool {
	return (err == nil && result == nil) || (err != nil && !failureModeDeny)
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Error("Error Occured", rz.Err(err))
	http.Error(res, "Something bad happened", http.StatusBadGateway)
}

func cacheResponse(ctx context.Context, requestId string, repo storage.Repository, metrics *metrics.MetricCollector, failureModeDeny bool) func(*http.Response) error {
	return func(response *http.Response) error {
		log.Debug("Got response from downstream service")
		metrics.DownstreamHit(response.StatusCode, response.Request.Method)

		if response.StatusCode >= 500 {
			log.Warn("Won't cache 5XX downstream responses")
			return nil
		}

		body, err := io.ReadAll(response.Body)
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

		if err = repo.Store(ctx, requestId, &storage.Response{Body: body, Header: header, Status: statusCode}); err != nil {
			if failureModeDeny {
				return err
			}
		}

		return nil
	}
}

func getRequestId(headerKeys []string, req *http.Request) (string, error) {
	var maybeRequestId string
	for _, key := range headerKeys {
		maybeRequestId = req.Header.Get(key)
		if maybeRequestId != "" {
			break
		}
	}

	if maybeRequestId == "" {
		return "", errors.New("RequestId header missing")
	}

	return maybeRequestId, nil
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

func IdempotencyHandler(repo storage.Repository, downstreamURL *url.URL) http.HandlerFunc {
	metrics := metrics.NewMetricCollector()

	proxy := httputil.NewSingleHostReverseProxy(downstreamURL)
	proxy.ErrorHandler = errHandler

	serverConfig := config.New().ServerConfig

	return func(res http.ResponseWriter, req *http.Request) {
		if utils.VariableMatchesRegexIn(req.URL.Path, serverConfig.PassthroughEndpoints) {
			log.Info(fmt.Sprintf("%v is a passthrough endpoint.", req.URL.Path))
			proxy.ServeHTTP(res, req)
			return
		}

		ctx := req.Context()
		requestId, err := getRequestId(serverConfig.IdempotencyKeys, req)

		if err != nil && serverConfig.FailureModeDeny {
			writeErrorResponse(res, http.StatusBadRequest, fmt.Sprintf("missing header(s) in HTTP request. Required headers: %v", serverConfig.IdempotencyKeys))
			return
		}

		result, err := repo.LookUp(ctx, requestId)

		if err != nil && serverConfig.FailureModeDeny {
			writeErrorResponse(res, http.StatusBadGateway, fmt.Sprintf("Storage did not respond in time or error occured: %v", err))
			return
		}

		if shouldProxyRequest(err, result, serverConfig.FailureModeDeny) {
			proxy.ModifyResponse = cacheResponse(ctx, requestId, repo, metrics, serverConfig.FailureModeDeny)
			log.Debug("response from downstream cached", rz.String("request-id", requestId))
			proxy.ServeHTTP(res, req)
			return
		}

		log.Info("serving request from memory", rz.String("request-id", requestId))
		metrics.CacheHit(result.Status, req.Method)
		serveResponseFromMemory(res, result)
	}
}
