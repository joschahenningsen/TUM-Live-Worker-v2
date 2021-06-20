package cfg

import "os"

var (
	WorkerID   = os.Getenv("workerID")
	TempDir    = "/recordings"                         // recordings will end up here before they are converted
	StorageDir = "/srv/cephfs/livestream/rec/TUM-Live" // recordings will end up here after they are converted
	IngestBase = os.Getenv("IngestBase")
	LrzUser    = os.Getenv("LrzUser")
	LrzMail    = os.Getenv("LrzMail")
	LrzPhone   = os.Getenv("LrzPhone")
	LrzSubDir  = os.Getenv("LrzSubDir")
)
