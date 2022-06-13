.PHONY : build run fresh test clean
include dev.env
export

BIN := cache-offloader

HASH := $(shell git rev-parse --short HEAD)
COMMIT_DATE := $(shell git show -s --format=%ci ${HASH})
BUILD_DATE := $(shell date '+%Y-%m-%d %H:%M:%S')
VERSION := ${HASH} (${COMMIT_DATE})


clean:
	go clean
	go clean -testcache
	rm -rf ./bin ./vendor go.sum coverage.out coverage.out.tmp
	go mod tidy

lint:
	golangci-lint run

format:
	gofumpt -l -w .

build:
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -o bin/${BIN} -ldflags="-s -w -X 'main.version=${VERSION}' -X 'main.buildDate=${BUILD_DATE}'" -v ./cmd/${BIN}.go 
	env GOARCH=arm64 GOOS=darwin go build -o bin/${BIN}-mac -ldflags="-s -w" -v ./cmd/${BIN}.go

run: fresh
	./bin/${BIN}-mac

fresh: clean build


test:
	go test -race -count=1 -v --coverprofile=coverage.txt ./...
	go tool cover -func coverage.txt | grep total

cov: test
	go tool cover -html=coverage.txt

.PHONY: setup
setup:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s v1.46.2