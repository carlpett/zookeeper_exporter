FROM scratch
ADD .build/zookeeper_exporter /
ENTRYPOINT ["/zookeeper_exporter"]
