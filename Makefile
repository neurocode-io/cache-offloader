.PHONY : build run fresh test clean
include dev.env
export

BIN := cache-offloader

HASH := $(shell git rev-parse --short HEAD)
COMMIT_DATE := $(shell git show -s --format=%ci ${HASH})
BUILD_DATE := $(shell date '+%Y-%m-%d %H:%M:%S')
VERSION := ${HASH} (${COMMIT_DATE})


deps:
	go mod download

build:
	go build -o app -ldflags="-X 'main.buildVersion=${VERSION}' -X 'main.buildDate=${BUILD_DATE}'" -v ./cmd/${BIN}.go

run: fresh
	./app

fresh: clean build


test: clean
	go test -race ./... -v

cov:
	go test ./... -v -coverprofile cp.out

html-cov: cov
	go tool cover -html=cp.out

clean:
	go clean
	go clean -testcache
	- rm -f app
