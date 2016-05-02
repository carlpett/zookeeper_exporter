package main

import (
  "bufio"
  "bytes"
  "net"
  "regexp"
  "strings"

  log "github.com/Sirupsen/logrus"
)

func fetchMetrics() ([]byte, bool) {
  log.Info("Fetching metrics from Zookeeper")

  data, ok := sendZkCommand("mntr")

  if !ok {
    log.Error("Failed to fetch metrics")
    return nil, false
  }

  buffer := bytes.Buffer {}
  for _, line := range strings.Split(data, "\n") {
    label := strings.Split(line, "\t")[0]
    stat, ok := stats[label]
    if ok {
      buffer.WriteString(replace(stat.Pattern, line, stat.Template))
    }
  }

  return buffer.Bytes(), true
}
func resetStatistics() {
  log.Info("Resetting Zookeeper statistics")
  _, ok := sendZkCommand("srst")
  if !ok {
    log.Warning("Failed to reset statistics")
  }
}

func replace(pattern string, source string, template string) string {
  r, _ := regexp.Compile(pattern)
  m := r.FindStringSubmatchIndex(source)

  res := []byte{}
  return string(r.ExpandString(res, template, source, m))
}

func sendZkCommand(fourLetterWord string) (string, bool) {
  log.Debugf("Connecting to Zookeeper at %s", *zookeeperAddr)

  conn, err := net.Dial("tcp", *zookeeperAddr)
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
