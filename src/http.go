package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type repository interface {
	lookUp(string) (http.Response, error)
	store(string, http.Response) error
}

func errHandler(res http.ResponseWriter, req *http.Request, err error) {
	log.Printf("Error occured: %v", err)
	http.Error(res, "Something bad happened", http.StatusBadGateway)
}

func responseHandler(res *http.Response) error {
	log.Printf("Got response %v", res)
	// TODO save the response to redis with requestId as key
	return nil
}

func livenessHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "Alive")
}

func readinessHandler(res http.ResponseWriter, req *http.Request) {
	// TODO add redis check
	fmt.Fprintf(res, "Alive")
}

func getRequestId(req *http.Request) (string, error) {
	maybeRequestId := req.Header.Get("x-request-id")

	if maybeRequestId == "" {
		return "", errors.New("RequestId header missing")
	}

	return maybeRequestId, nil
}

func indempotencyHandler(r repository, downstreamURL *url.URL, allowedEndpoints []string) http.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(downstreamURL)

	proxy.ErrorHandler = errHandler
	proxy.ModifyResponse = responseHandler

	return func(res http.ResponseWriter, req *http.Request) {
		// fdont do anything to the endpoints not in the allowedEndpoints list
		if variableMatchesRegexIn(req.URL.Path, allowedEndpoints) == false {
			proxy.ServeHTTP(res, req)
			return
		}

		// only cache the calls that are configured in the allowedEndpoints list
		requestId, err := getRequestId(req)
		if err != nil {
			// TODY check what to do when header is missing
			log.Println("missing requestId header")
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := r.lookUp(requestId)
		if err != nil {
			proxy.ServeHTTP(res, req)
			return
		}

		log.Printf("serving from memory requestId %v", requestId)
		log.Printf("serving from memory requestId %v", result)

	}
}
