package main

import (
  log "github.com/Sirupsen/logrus"
  "regexp"
  "strings"
  "gopkg.in/yaml.v2"
  "github.com/jasonlvhit/gocron"
  "io/ioutil"
  "net"
  "net/http"
  "bufio"
  "bytes"
  "os"
  "os/signal"
  "syscall"
)

type Stat struct {
  Pattern string  `yaml:"pattern"`
  Template string `yaml:"template"`
}
var stats map[string] Stat = loadStatsDefinitions()
var formattedMetrics []byte

func metricsHandler(w http.ResponseWriter, r *http.Request) {
  log.WithFields(log.Fields { "request": r }).Debug("Incoming request to /metrics")

  if formattedMetrics == nil {
    log.Error("Call to /metrics before successfully collecting metrics from Zookeeper!")

    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write(formattedMetrics)
}

func main() {
  log.SetLevel(log.DebugLevel)
  log.Info("Starting zookeeper_exporter")

  go schedule()
  go serveMetrics()

  exitChannel := make(chan os.Signal)
  signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
  exitSignal := <- exitChannel
  log.WithFields(log.Fields { "signal": exitSignal }).Infof("Caught %s signal, exiting", exitSignal)
}

func serveMetrics() {
  log.Info("Starting metric http endpoint on :8090")
  http.HandleFunc("/metrics", metricsHandler)
  log.Fatal(http.ListenAndServe(":8090", nil))  
}

func schedule() {
  log.Info("Starting scheduler")
  gocron.Every(10).Seconds().Do(updateMetrics)
  gocron.Every(1).Hour().Do(resetStatistics)
  <- gocron.Start()
}

func updateMetrics() {
  log.Info("Updating metrics from Zookeeper")

  data, ok := sendZkCommand("mntr")

  if !ok {
    log.Error("Failed to update metrics")
    return
  }

  buffer := bytes.Buffer {}
  for _, line := range strings.Split(data, "\n") {
    label := strings.Split(line, "\t")[0]
    stat, ok := stats[label]
    if ok {
      buffer.WriteString(replace(stat.Pattern, line, stat.Template))
    }
  }

  formattedMetrics = buffer.Bytes()
}
func resetStatistics() {
  log.Info("Resetting Zookeeper statistics")
  _, ok := sendZkCommand("srst")
  if !ok {
    log.Warning("Failed to reset statistics")
  }
}

func sendZkCommand(fourLetterWord string) (string, bool) {
  log.Debug("Connecting to Zookeeper at localhost:2181")

  conn, err := net.Dial("tcp", "localhost:2181")
  if err != nil {
    log.WithFields(log.Fields { "error": err }).Error("Unable to open connection to Zookeeper")
    return "", false
  }
  defer conn.Close()

  log.WithFields(log.Fields {"command": fourLetterWord}).Debug("Sending four letter word")
  conn.Write([]byte(fourLetterWord))
  scanner := bufio.NewScanner(conn)

  buffer := bytes.Buffer {}
  for scanner.Scan() {
    buffer.WriteString(scanner.Text() + "\n")
  }
  if err = scanner.Err(); err != nil {
    log.WithFields(log.Fields { "error": err }).Error("Error parsing response from Zookeeper")
    return "", false
  }
  log.Debug("Successfully retrieved reply")

  return buffer.String(), true
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

func replace(pattern string, foobar string, template string) string {
  r, _ := regexp.Compile(pattern)
  m := r.FindStringSubmatchIndex(foobar)

  res := []byte{}
  return string(r.ExpandString(res, template, foobar, m))
}