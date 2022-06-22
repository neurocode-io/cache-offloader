# cache-offloader
[![CICD](https://github.com/neurocode-io/cache-offloader/actions/workflows/main.yml/badge.svg)](https://github.com/neurocode-io/cache-offloader/actions/workflows/main.yml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/gomods/athens)


[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=alert_status)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=security_rating)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)


# Install

pre go 1.15
go get github.com/neurocode-io/cache-offloader@v0.1.1

after go 1.15
go install github.com/neurocode-io/cache-offloader@v0.1.1


# Using the service

This service provides a stale while revalidate cache for your HTTP requests. If a request is made to a URL that is already in the cache, the cache will be used. The cache will also be used if the cache is stale. However, an asynchronous request will be made to the server to refresh the cache if the cache is stale.

Check out the `dev.env` file for configuration options.

Essentially you can choose between in-memory or redis persistence. 

Redis persistence is recommended when you have multiple processes running the service (or multiple pods).


The environment variable CACHE_IGNORE_ENDPOINTS defines the endpoints that are allowed to passthrough without any checks from this service.

You can also choose to incude query parameters in the cache key. `CACHE_SHOULD_HASH_QUERY` is a boolean flag that defines whether or not to hash the query parameters. 

By `CACHE_HASH_QUERY_IGNORE` you can define which query parameters should be ignored when hashing the query.

Another useful configuration option is `CACHE_STALE_WHILE_REVALIDATE_SEC`. This defines the time in seconds that the cache entry is considered stale and the cache is revalidated asynchronously.


# Dev

```
make run
```

Check its working:

```
http localhost:8000/probes/readiness
```

# Test

```
go install github.com/golang/mock/mockgen@latest
```
