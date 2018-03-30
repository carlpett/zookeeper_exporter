# ref: https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#use-multi-stage-builds
FROM golang:1.9.2-alpine3.6 AS build

# Install tools required to build the project
# We need to run `docker build --no-cache .` to update those dependencies
RUN apk add --no-cache git
RUN go get github.com/golang/dep/cmd/dep

# Gopkg.toml and Gopkg.lock lists project dependencies
# These layers are only re-built when Gopkg files are updated
COPY Gopkg.lock Gopkg.toml /go/src/zookeeper-exporter/
WORKDIR /go/src/zookeeper-exporter/
# Install library dependencies
RUN dep ensure -vendor-only

# Copy all project and build it
# This layer is rebuilt when ever a file has changed in the project directory
COPY . /go/src/zookeeper-exporter/
RUN go build -o /bin/zookeeper-exporter

# This results in a single layer image
FROM alpine:3.6
COPY --from=build /bin/zookeeper-exporter /bin/zookeeper-exporter
ENTRYPOINT ["/bin/zookeeper-exporter"]
