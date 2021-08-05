FROM golang:1.16-buster as build

WORKDIR /go/src/app
COPY . ./

RUN go mod download \
  && go build -o /go/bin/app

FROM gcr.io/distroless/base-debian10

COPY --from=build /go/bin/app /

CMD ["/app"]