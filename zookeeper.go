package main

import (
  "bufio"
  "bytes"
  "net"
  "regexp"
  "strings"

  log "github.com/Sirupsen/logrus"
)

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

func replace(pattern string, foobar string, template string) string {
  r, _ := regexp.Compile(pattern)
  m := r.FindStringSubmatchIndex(foobar)

  res := []byte{}
  return string(r.ExpandString(res, template, foobar, m))
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
