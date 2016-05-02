package main

import (
  "net/http"

  log "github.com/Sirupsen/logrus"
)

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

func metricsHandler(w http.ResponseWriter, r *http.Request) {
  log.WithFields(log.Fields { "request": r }).Debugf("Serving metrics request")

  metrics, ok := fetchMetrics()

  if !ok {
    log.Warning("Scrape failed")
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Internal server error"))
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write(metrics)

  // Reset stats after each scrape, since it is the only way of getting something at least almost consistent
  resetStatistics()
}

func serveMetrics() {
  log.Infof("Starting metric http endpoint on %s", *bindAddr)
  http.HandleFunc(*metricsPath, metricsHandler)
  http.HandleFunc("/", rootHandler)
  log.Fatal(http.ListenAndServe(*bindAddr, nil))
}
