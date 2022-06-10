# cache-offloader
[![CICD](https://github.com/neurocode-io/cache-offloader/actions/workflows/main.yml/badge.svg)](https://github.com/neurocode-io/cache-offloader/actions/workflows/main.yml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/gomods/athens)


[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=alert_status)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=security_rating)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=neurocode-io_cache-offloader&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=neurocode-io_cache-offloader)

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


# Configuration
... ADD MORE

CACHE_IGNORE_ENDPOINTS=/management/prometheus, /api-docs

Note that the environment variable CACHE_IGNORE_ENDPOINTS defines the endpoints that are allowed to passthrough without any checks from this service.
