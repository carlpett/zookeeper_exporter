BUILDINFO = $(subst ${GOPATH}/src/,,${PWD})/vendor/github.com/prometheus/common/version

VERSION  = $(shell git describe --always --tags --dirty=-dirty)
REVISION = $(shell git rev-parse HEAD)
BRANCH   = $(shell git rev-parse --abbrev-ref HEAD)

DOCKER_REPO       = carlpett
DOCKER_IMAGE_NAME = zookeeper_exporter
DOCKER_IMAGE_TAG  = ${VERSION}

all: build

build:
	@echo ">> building zookeeper_exporter"
	@CGO_ENABLED=0 go build -ldflags "\
            -X ${BUILDINFO}.Version=${VERSION} \
            -X ${BUILDINFO}.Revision=${REVISION} \
            -X ${BUILDINFO}.Branch=${BRANCH} \
            -X ${BUILDINFO}.BuildUser=$(USER)@$(HOSTNAME) \
            -X ${BUILDINFO}.BuildDate=$(shell date +%Y-%m-%dT%T%z)"

docker:
	@echo ">> building docker image ${DOCKER_REPO}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
	@docker build -t ${DOCKER_REPO}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .

release: build bin/github-release
	@echo ">> uploading release ${VERSION}"
	@./bin/github-release upload -t ${VERSION} -n zookeeper_exporter -f zookeeper_exporter

bin/github-release:
	@mkdir -p bin
	@curl -sL 'https://github.com/aktau/github-release/releases/download/v0.6.2/linux-amd64-github-release.tar.bz2' | tar xjf - --strip-components 3 -C bin

.PHONY: all build docker release
