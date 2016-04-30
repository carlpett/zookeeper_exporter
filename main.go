package main

import (
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
var formattedMetrics []byte

func main() {
  log.SetLevel(log.DebugLevel)
  log.Info("Starting zookeeper_exporter")

  go scheduleTasks()
  go serveMetrics()

  exitChannel := make(chan os.Signal)
  signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
  exitSignal := <- exitChannel
  log.WithFields(log.Fields { "signal": exitSignal }).Infof("Caught %s signal, exiting", exitSignal)
}

func scheduleTasks() {
  log.Info("Starting scheduler")
  gocron.Every(10).Seconds().Do(updateMetrics)
  gocron.Every(1).Hour().Do(resetStatistics)
  <- gocron.Start()
}

func loadStatsDefinitions() map[string] Stat {
  log.WithFields(log.Fields {"file": "stats.yml"}).Debug("Loading parser definitions")
  data, err := ioutil.ReadFile("stats.yml")
  if err != nil {
    log.WithFields(log.Fields { "error": err, "file": "stats.yml" }).Fatal("Unable to load parser definitions")
  }

  stats := make(map[string]Stat)
  err = yaml.Unmarshal(data, &stats)
  if err != nil {
    log.WithFields(log.Fields { "error": err }).Fatal("Could not parse parser definitions")
  }

  return stats
}
