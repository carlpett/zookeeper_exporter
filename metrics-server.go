package main

import (
  "net/http"

  log "github.com/Sirupsen/logrus"
)

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

func serveMetrics() {
  log.Info("Starting metric http endpoint on :8090")
  http.HandleFunc("/metrics", metricsHandler)
  log.Fatal(http.ListenAndServe(":8090", nil))  
}
