package main

import (
	"github.com/joschahenningsen/TUM-Live-Worker-v2/api"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true}) // log with time

	// setup api
	api.InitApi(":8082")
}
