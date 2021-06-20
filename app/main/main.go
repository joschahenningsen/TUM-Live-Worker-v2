package main

import (
	"github.com/joschahenningsen/TUM-Live-Worker-v2/api"
	log "github.com/sirupsen/logrus"
)

func main() {
	// log with time, fmt "23.09.2021 10:00:00"
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "02.01.2006 15:04:05", FullTimestamp: true})

	// setup api
	api.InitApi(":8082")
}
