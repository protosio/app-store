FROM golang:1.10

ADD . "/go/src/app-store/"
WORKDIR "/go/src/app-store"
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o app-store *.go
RUN chmod +x /go/src/app-store/app-store

ENTRYPOINT ["/go/src/app-store/app-store"]
