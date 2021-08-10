package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"context"

	"dpd.de/indempotency-offloader/config"
	"dpd.de/indempotency-offloader/pkg/storage"
	"dpd.de/indempotency-offloader/pkg/utils"
)

func shouldProxyRequest(err error, result *http.Response, failureModeDeny bool) bool {
	// if err and result = nil storage did not find the key thus we should forward the request
	// if we allow for storage errors (err and failureModeDeny) we should also forward the request
	return (err == nil && result == nil) || (err != nil && failureModeDeny == false)
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Printf("Error occured: %v", err)
	http.Error(res, "Something bad happened", http.StatusBadGateway)
}

func cacheResponse(ctx context.Context, requestId string, repo storage.Repository) func(*http.Response) error {
	return func(downstream *http.Response) error {
		log.Printf("Got response from downstream service %v", downstream)

		serializedResp, err := httputil.DumpResponse(downstream, true)
		if err != nil {
			return err
		}

		if err := repo.Store(ctx, requestId, serializedResp); err != nil {
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

func IndempotencyHandler(repo storage.Repository, downstreamURL *url.URL) http.HandlerFunc {
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
		requestId, err := getRequestId(serverConfig.IndempotencyKeys, req)
		if err != nil && serverConfig.FailureModeDeny {
			log.Printf("missing %v headers", serverConfig.IndempotencyKeys)
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Sprintf("missing %v headers in HTTP request", serverConfig.IndempotencyKeys)))
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

		res.WriteHeader(result.StatusCode)
		json.NewEncoder(res).Encode(result.Body)

	}
}
