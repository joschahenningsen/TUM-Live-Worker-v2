package cfg

import (
	"os"
)

var (
	WorkerID     string
	TempDir      string // recordings will end up here before they are converted
	StorageDir   string // recordings will end up here after they are converted
	IngestBase   string
	LrzUser      string
	LrzMail      string
	LrzPhone     string
	LrzSubDir    string
	MainBase     string
	LrzUploadUrl string
	LogDir       string
)

// init stops the execution if any of the required config variables are unset.
func init() {
	WorkerID = os.Getenv("WorkerID")
	TempDir = "/recordings"                            // recordings will end up here before they are converted
	StorageDir = "/srv/cephfs/livestream/rec/TUM-Live" // recordings will end up here after they are converted
	IngestBase = os.Getenv("IngestBase")
	LrzUser = os.Getenv("LrzUser")
	LrzMail = os.Getenv("LrzMail")
	LrzPhone = os.Getenv("LrzPhone")
	LrzSubDir = os.Getenv("LrzSubDir")
	LrzUploadUrl = os.Getenv("LrzUploadUrl")
	MainBase = os.Getenv("MainBase") // eg. live.mm.rbg.tum.de
	LogDir = os.Getenv("LogDir")
}
