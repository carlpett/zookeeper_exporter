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
  log.WithFields(log.Fields { "request": r }).Debug("Incoming request to /metrics")

  if formattedMetrics == nil {
    log.Error("Call to /metrics before successfully collecting metrics from Zookeeper!")

    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Internal server error"))
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write(formattedMetrics)
}

func serveMetrics() {
  log.Info("Starting metric http endpoint on :8090")
  http.HandleFunc("/metrics", metricsHandler)
  http.HandleFunc("/", rootHandler)
  log.Fatal(http.ListenAndServe(":8090", nil))
}
