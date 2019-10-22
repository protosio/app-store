FROM golang:1.13.3 as builder

ADD . "/go/src/github.com/protosio/app-store"
WORKDIR "/go/src/github.com/protosio/app-store"
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.6.2/migrate.linux-amd64.tar.gz | tar xvz && mv migrate.linux-amd64 /usr/bin/migrate && chmod +x /usr/bin/migrate
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app-store main.go

FROM alpine:3.10.2
RUN apk add ca-certificates
COPY --from=builder /go/src/github.com/protosio/app-store/app-store /usr/bin/
COPY --from=builder /go/src/github.com/protosio/app-store/migrations /root/migrations/
COPY --from=builder /usr/bin/migrate /usr/bin/
RUN chmod +x /usr/bin/app-store /usr/bin/migrate

ENTRYPOINT ["/usr/bin/app-store"]
