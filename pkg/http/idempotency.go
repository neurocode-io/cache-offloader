package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"context"

	"dpd.de/idempotency-offloader/config"
	"dpd.de/idempotency-offloader/pkg/storage"
	"dpd.de/idempotency-offloader/pkg/utils"
)

func shouldProxyRequest(err error, result *storage.Response, failureModeDeny bool) bool {
	// if err and result = nil storage did not find the key thus we should forward the request
	// if we allow for storage errors (err and failureModeDeny) we should also forward the request
	return (err == nil && result == nil) || (err != nil && failureModeDeny == false)
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Printf("Error occured: %v", err)
	http.Error(res, "Something bad happened", http.StatusBadGateway)
}

func cacheResponse(ctx context.Context, requestId string, repo storage.Repository) func(*http.Response) error {
	return func(response *http.Response) error {
		log.Printf("Got response from downstream service %v", response)
		if response.StatusCode >= 500 {
			log.Printf("Won't cache 5XX downstream responses")
			return nil
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		header := response.Header
		newBody := ioutil.NopCloser(bytes.NewReader(body))

		response.Body = newBody

		if err = repo.Store(ctx, requestId, &storage.Response{Body: body, Header: header}); err != nil {
			return err
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

func IdempotencyHandler(repo storage.Repository, downstreamURL *url.URL) http.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(downstreamURL)
	serverConfig := config.New().ServerConfig

	proxy.ErrorHandler = errHandler

	return func(res http.ResponseWriter, req *http.Request) {
		// proxy the requests not in the allowedEndpoints list
		if utils.VariableMatchesRegexIn(req.URL.Path, serverConfig.AllowedEndpoints) == false {
			proxy.ServeHTTP(res, req)
			return
		}

		ctx := req.Context()
		requestId, err := getRequestId(serverConfig.IdempotencyKeys, req)
		if err != nil && serverConfig.FailureModeDeny {
			log.Printf("missing %v headers", serverConfig.IdempotencyKeys)
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Sprintf("missing %v headers in HTTP request", serverConfig.IdempotencyKeys)))
			return
		}

		result, err := repo.LookUp(ctx, requestId)
		if shouldProxyRequest(err, result, serverConfig.FailureModeDeny) {
			// initialize proxyResponse callback
			proxy.ModifyResponse = cacheResponse(ctx, requestId, repo)
			proxy.ServeHTTP(res, req)

			return
		}

		if err != nil && serverConfig.FailureModeDeny {
			log.Printf("Storage did not respond in time or error occured: %v", err)
			res.WriteHeader(http.StatusBadGateway)
			res.Write([]byte("Storage did not respond in time or error occured"))
			return
		}

		log.Printf("serving from memory requestId %v", requestId)

		for key, values := range result.Header {
			res.Header()[key] = values
		}

		res.Write(result.Body)
	}
}
