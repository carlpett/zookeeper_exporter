package main

import (
	"bufio"
	"bytes"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

type zookeeperCollector struct {
	upIndicator *prometheus.Desc
	metrics     map[string]zookeeperMetric
}
type zookeeperMetric struct {
	desc          *prometheus.Desc
	extract       func(string) float64
	extractLabels func(s string) []string
	valType       prometheus.ValueType
}

func init() {
	prometheus.MustRegister(NewZookeeperCollector())
}

func parseFloatOrZero(s string) float64 {
	res, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Warningf("Failed to parse to float64: %s", err)
		return 0.0
	}
	return res
}
func NewZookeeperCollector() *zookeeperCollector {
	return &zookeeperCollector{
		upIndicator: prometheus.NewDesc("zk_up", "Exporter successful", nil, nil),
		metrics: map[string]zookeeperMetric{
			"zk_avg_latency": {
				desc:    prometheus.NewDesc("zk_avg_latency", "Average latency of requests", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_max_latency": {
				desc:    prometheus.NewDesc("zk_max_latency", "Maximum seen latency of requests", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_min_latency": {
				desc:    prometheus.NewDesc("zk_min_latency", "Minimum seen latency of requests", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_packets_received": {
				desc:    prometheus.NewDesc("zk_packets_received", "Number of packets received", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.CounterValue,
			},
			"zk_packets_sent": {
				desc:    prometheus.NewDesc("zk_packets_sent", "Number of packets sent", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.CounterValue,
			},
			"zk_num_alive_connections": {
				desc:    prometheus.NewDesc("zk_num_alive_connections", "Number of active connections", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_outstanding_requests": {
				desc:    prometheus.NewDesc("zk_outstanding_requests", "Number of outstanding requests", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_server_state": {
				desc:    prometheus.NewDesc("zk_server_state", "Server state (leader/follower)", []string{"state"}, nil),
				extract: func(s string) float64 { return 1 },
				extractLabels: func(s string) []string {
					return []string{s}
				},
				valType: prometheus.UntypedValue,
			},
			"zk_znode_count": {
				desc:    prometheus.NewDesc("zk_znode_count", "Number of znodes", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_watch_count": {
				desc:    prometheus.NewDesc("zk_watch_count", "Number of watches", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_ephemerals_count": {
				desc:    prometheus.NewDesc("zk_ephemerals_count", "Number of ephemeral nodes", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_approximate_data_size": {
				desc:    prometheus.NewDesc("zk_approximate_data_size", "Approximate size of data set", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_open_file_descriptor_count": {
				desc:    prometheus.NewDesc("zk_open_file_descriptor_count", "Number of open file descriptors", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_max_file_descriptor_count": {
				desc:    prometheus.NewDesc("zk_max_file_descriptor_count", "Maximum number of open file descriptors", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.CounterValue,
			},
			"zk_followers": {
				desc:    prometheus.NewDesc("zk_followers", "Number of followers", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_synced_followers": {
				desc:    prometheus.NewDesc("zk_synced_followers", "Number of followers in sync", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
			"zk_pending_syncs": {
				desc:    prometheus.NewDesc("zk_pending_syncs", "Number of followers with syncronizations pending", nil, nil),
				extract: func(s string) float64 { return parseFloatOrZero(s) },
				valType: prometheus.GaugeValue,
			},
		},
	}
}

func (c *zookeeperCollector) Describe(ch chan<- *prometheus.Desc) {
	log.Debugf("Sending %d metrics descriptions", len(c.metrics))
	for _, i := range c.metrics {
		ch <- i.desc
	}
}

func (c *zookeeperCollector) Collect(ch chan<- prometheus.Metric) {
	log.Info("Fetching metrics from Zookeeper")

	data, ok := sendZkCommand("mntr")

	if !ok {
		log.Error("Failed to fetch metrics")
		ch <- prometheus.MustNewConstMetric(c.upIndicator, prometheus.GaugeValue, 0)
		return
	}

	data = strings.TrimSpace(data)
	status := 1.0
	for _, line := range strings.Split(data, "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			log.WithFields(log.Fields{"data": line}).Warn("Unexpected format of returned data, expected tab-separated key/value.")
			status = 0
			continue
		}
		label, value := parts[0], parts[1]
		metric, ok := c.metrics[label]
		if ok {
			log.Debugf("Sending metric %s=%s", label, value)
			if metric.extractLabels != nil {
				ch <- prometheus.MustNewConstMetric(metric.desc, metric.valType, metric.extract(value), metric.extractLabels(value)...)
			} else {
				ch <- prometheus.MustNewConstMetric(metric.desc, metric.valType, metric.extract(value))
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(c.upIndicator, prometheus.GaugeValue, status)

	if *resetOnScrape {
		resetStatistics()
	}
}
func resetStatistics() {
	log.Info("Resetting Zookeeper statistics")
	_, ok := sendZkCommand("srst")
	if !ok {
		log.Warning("Failed to reset statistics")
	}
}

const (
	timeoutSeconds = 5
)

func sendZkCommand(fourLetterWord string) (string, bool) {
	log.Debugf("Connecting to Zookeeper at %s", *zookeeperAddr)

	conn, err := net.Dial("tcp", *zookeeperAddr)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to open connection to Zookeeper")
		return "", false
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(timeoutSeconds * time.Second))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to set timeout on Zookeeper connection")
		return "", false
	}

	log.WithFields(log.Fields{"command": fourLetterWord}).Debug("Sending four letter word")
	_, err = conn.Write([]byte(fourLetterWord))
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error sending command to Zookeeper")
		return "", false
	}
	scanner := bufio.NewScanner(conn)

	buffer := bytes.Buffer{}
	for scanner.Scan() {
		buffer.WriteString(scanner.Text() + "\n")
	}
	if err = scanner.Err(); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error parsing response from Zookeeper")
		return "", false
	}
	log.Debug("Successfully retrieved reply")

	return buffer.String(), true
}
