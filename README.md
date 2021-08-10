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


FAILURE_MODE_DENY=true 

If the header is missing and failure_mode_deny is set to true a 400 response is returned
If there is an error in calling redis or redis returns an error and failure_mode_deny is set to true, a 504 response is returned.
