FROM golang:1.10 AS builder
WORKDIR /go/src/github.com/carlpett/zookeeper_exporter/
COPY . .
RUN make build

FROM scratch
EXPOSE 9141
USER 1000
ENTRYPOINT ["/zookeeper_exporter"]
COPY --from=builder /go/src/github.com/carlpett/zookeeper_exporter/zookeeper_exporter /zookeeper_exporter
