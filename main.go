package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
)

func init() {
	flag.Parse()

	parsedLevel, err := log.ParseLevel(*rawLevel)
	if err != nil {
		log.Fatal(err)
	}
	logLevel = parsedLevel

	if *enableTLS && (*certPath == "" || *certKeyPath == "") {
		log.Fatal("-enable-tls requires -cert and -cert-key")
	}

	if *logJSON {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	}

	prometheus.MustRegister(version.NewCollector("zookeeper_exporter"))
}

var (
	logLevel      log.Level = log.InfoLevel
	logJSON                 = flag.Bool("log-json", false, "Log output as JSON")
	bindAddr                = flag.String("bind-addr", ":9141", "bind address for the metrics server")
	enableTLS               = flag.Bool("enable-tls", false, "Connect to zookeeper using TLS. Requires -cert and -cert-key")
	certPath                = flag.String("cert", "", "path to certificate including any intermediaries")
	certKeyPath             = flag.String("cert-key", "", "path to certificate key")
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
	http.Handle(*metricsPath, promhttp.Handler())
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
