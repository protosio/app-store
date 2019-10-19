FROM golang:1.13.3 as builder

ADD . "/go/src/github.com/protosio/app-store"
WORKDIR "/go/src/github.com/protosio/app-store"
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app-store main.go

FROM alpine:latest
RUN apk add ca-certificates
COPY --from=builder /go/src/github.com/protosio/app-store/app-store /usr/bin/
RUN chmod +x /usr/bin/app-store

ENTRYPOINT ["/usr/bin/app-store"]
