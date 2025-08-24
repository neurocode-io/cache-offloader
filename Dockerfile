FROM gcr.io/distroless/base-debian12:nonroot

COPY bin/cache-offloader /usr/local/bin/

ENTRYPOINT ["cache-offloader"]