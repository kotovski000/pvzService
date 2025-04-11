FROM golang:1.23.0

WORKDIR ${GOPATH}/pvz-shop/
COPY . ${GOPATH}/pvz-shop/

RUN go mod download


RUN go build -o /build ./cmd \
    && go clean -cache -modcache

EXPOSE 8080 9000

CMD ["/build"]