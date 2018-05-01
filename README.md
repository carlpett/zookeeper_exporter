# zookeeper_exporter [![CircleCI](https://circleci.com/gh/carlpett/zookeeper_exporter.svg?style=shield)](https://circleci.com/gh/carlpett/zookeeper_exporter) [![DockerHub](https://img.shields.io/docker/build/carlpett/zookeeper_exporter.svg?style=shield)](https://hub.docker.com/r/carlpett/zookeeper_exporter/)

A Prometheus exporter for Zookeeper 3.4+. It send the `mntr` command to a Zookeeper node and converts the output to Prometheus format. 

## Usage
Download the [latest release](https://github.com/carlpett/zookeeper_exporter/releases), pull [the Docker image](https://hub.docker.com/r/carlpett/zookeeper_exporter/) or follow the instructions below for building the source.

There is a `-help` flag for listing the available flags.

## Building from source
`go get -u github.com/carlpett/zookeeper_exporter` and then `make build`.

## Limitations
Due to the type of data exposed by Zookeeper's `mntr` command, it currently resets Zookeeper's internal statistics every time it is scraped. This makes it unsuitable for having multiple parallel scrapers.
