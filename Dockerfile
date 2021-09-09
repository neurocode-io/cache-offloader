FROM golang:1.16-buster as build

WORKDIR /go/src/app
COPY . ./

RUN go mod download \
  && make build

FROM gcr.io/distroless/base-debian10

COPY --from=build /go/src/app/app /

CMD ["/app"]