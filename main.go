package main

import (
  "flag"
  "io/ioutil"
  "os"
  "os/signal"
  "syscall"

  "gopkg.in/yaml.v2"
  "github.com/jasonlvhit/gocron"
  log "github.com/Sirupsen/logrus"
)

type Stat struct {
  Pattern string  `yaml:"pattern"`
  Template string `yaml:"template"`
}
var stats map[string] Stat = loadStatsDefinitions()

func init() {
  flag.Parse()

  parsedLevel, err := log.ParseLevel(*rawLevel)
  if err != nil {
    log.Fatal(err)
  }
  logLevel = parsedLevel
}
var logLevel log.Level = log.InfoLevel
var bindAddr = flag.String("bind-addr", ":9141", "bind address for the metrics server")
var metricsPath = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
var zookeeperAddr = flag.String("zookeeper", "localhost:2181", "host:port for zookeeper socket")
var rawLevel = flag.String("log-level", "info", "log level")
var statsFile = flag.String("stats-file", "stats.yml", "yaml file containing stats definitions")

func main() {
  log.SetLevel(logLevel)
  log.Info("Starting zookeeper_exporter")

  go scheduleReset()
  go serveMetrics()

  exitChannel := make(chan os.Signal)
  signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
  exitSignal := <- exitChannel
  log.WithFields(log.Fields { "signal": exitSignal }).Infof("Caught %s signal, exiting", exitSignal)
}

func scheduleReset() {
  log.Info("Scheduling hourly reset")
  gocron.Every(1).Hour().Do(resetStatistics)
  <- gocron.Start()
}

func loadStatsDefinitions() map[string] Stat {
  log.WithFields(log.Fields {"file": *statsFile}).Debug("Loading parser definitions")
  data, err := ioutil.ReadFile(*statsFile)
  if err != nil {
    log.WithFields(log.Fields { "error": err, "file": *statsFile }).Fatal("Unable to load parser definitions")
  }

  stats := make(map[string]Stat)
  err = yaml.Unmarshal(data, &stats)
  if err != nil {
    log.WithFields(log.Fields { "error": err }).Fatal("Could not parse parser definitions")
  }

  return stats
}
