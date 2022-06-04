FROM gcr.io/distroless/base-debian11

COPY bin/cache-offloader /usr/local/bin/

ENTRYPOINT ["cache-offloader"]