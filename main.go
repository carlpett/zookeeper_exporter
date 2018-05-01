package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

func init() {
	flag.Parse()

	parsedLevel, err := log.ParseLevel(*rawLevel)
	if err != nil {
		log.Fatal(err)
	}
	logLevel = parsedLevel

	prometheus.MustRegister(version.NewCollector("zookeeper_exporter"))
}

var (
	logLevel      log.Level = log.InfoLevel
	bindAddr                = flag.String("bind-addr", ":9141", "bind address for the metrics server")
	metricsPath             = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
	zookeeperAddr           = flag.String("zookeeper", "localhost:2181", "host:port for zookeeper socket")
	rawLevel                = flag.String("log-level", "info", "log level")
	resetOnScrape           = flag.Bool("reset-on-scrape", true, "should a reset command be sent to zookeeper on each scrape")
	showVersion             = flag.Bool("version", false, "show version and exit")
)

func main() {
	log.Info(version.Print("zookeeper_exporter"))
	log.SetLevel(logLevel)
	if *showVersion {
		return
	}

	log.Info("Starting zookeeper_exporter")

	go serveMetrics()

	exitChannel := make(chan os.Signal)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	exitSignal := <-exitChannel
	log.WithFields(log.Fields{"signal": exitSignal}).Infof("Caught %s signal, exiting", exitSignal)
}

func serveMetrics() {
	log.Infof("Starting metric http endpoint on %s", *bindAddr)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html>
		<head><title>Zookeeper Exporter</title></head>
		<body>
		<h1>Zookeeper Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
}
