FROM alpine:3.10 AS certs
RUN apk update \
    && apk add ca-certificates

FROM golang:1.10 AS builder
WORKDIR /go/src/github.com/carlpett/zookeeper_exporter/
COPY . .
RUN make build

FROM scratch
EXPOSE 9141
USER 1000
COPY --from=builder /go/src/github.com/carlpett/zookeeper_exporter/zookeeper_exporter /zookeeper_exporter
COPY --from=certs /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["/zookeeper_exporter"]
