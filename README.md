# idempotency-sidecar-offloader


# Dev

```
docker-compose up -d
make run
```

Check its working:

```
http localhost:8000/probes/readiness
```


# Configuration

FAILURE_MODE_DENY=true 

If the header is missing and failure_mode_deny is set to true a 400 response is returned
If there is an error in calling redis or redis returns an error and failure_mode_deny is set to true, a 504 response is returned.



DOWNSTREAM_PASSTHROUGH_ENDPOINTS=/management/prometheus, /api-docs

Note that the environment variable DOWNSTREAM_PASSTHROUGH_ENDPOINTS defines the endpoints that are allowed to passthrough without any checks from this service.

Note that the values in the list are greedy. Which means that for example /management/prometheus also matches /management/prometheus/a/b !