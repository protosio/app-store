FROM golang:1.10

ADD . "/go/src/github.com/protosio/app-store"
WORKDIR "/go/src/github.com/protosio/app-store"
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o app-store main.go
RUN chmod +x app-store

ENTRYPOINT ["/go/src/github.com/protosio/app-store/app-store"]
