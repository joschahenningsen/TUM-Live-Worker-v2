package cfg

import (
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	WorkerID   string
	TempDir    string // recordings will end up here before they are converted
	StorageDir string // recordings will end up here after they are converted
	IngestBase string
	LrzUser    string
	LrzMail    string
	LrzPhone   string
	LrzSubDir  string
	MainBase   string
)

// init stops the execution if any of the required config variables are unset.
func init() {
	if key, found := os.LookupEnv("WorkerID"); !found {
		log.Fatalln("env workerID not provided")
	} else {
		WorkerID = key
	}
	TempDir = "/recordings"                            // recordings will end up here before they are converted
	StorageDir = "/srv/cephfs/livestream/rec/TUM-Live" // recordings will end up here after they are converted
	IngestBase = os.Getenv("IngestBase")
	LrzUser = os.Getenv("LrzUser")
	LrzMail = os.Getenv("LrzMail")
	LrzPhone = os.Getenv("LrzPhone")
	LrzSubDir = os.Getenv("LrzSubDir")
	MainBase = os.Getenv("MainBase") // eg. live.mm.rbg.tum.de
	if WorkerID == "" {
		log.Fatalln("env workerID not provided")
	}
}
