# cache-offloader
[![CICD](https://github.com/neurocode-io/cache-offloader/actions/workflows/main.yml/badge.svg)](https://github.com/neurocode-io/cache-offloader/actions/workflows/main.yml)

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
