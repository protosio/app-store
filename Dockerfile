FROM golang:1.13.3

ADD . "/go/src/github.com/protosio/app-store"
WORKDIR "/go/src/github.com/protosio/app-store"
RUN go build -o app-store main.go
RUN chmod +x app-store

ENTRYPOINT ["/go/src/github.com/protosio/app-store/app-store"]
