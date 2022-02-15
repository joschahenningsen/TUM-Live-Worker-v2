package cfg

import (
	"context"
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"time"
)

var (
	WorkerID     string // authentication token, unique for every worker, used to verify all calls
	TempDir      string // recordings will end up here before they are converted
	StorageDir   string // recordings will end up here after they are converted
	LrzUser      string
	LrzMail      string
	LrzPhone     string
	LrzSubDir    string
	MainBase     string
	LrzUploadUrl string
	LogDir       string
	Hostname     string
	Token        string // setup token. Used to connect initially and to get a "WorkerID"
)

// init stops the execution if any of the required config variables are unset.
func init() {
	// JoinToken is required to join the main server as a worker
	Token = os.Getenv("Token")
	if Token == "" {
		log.Fatal("Environment variable Token is not set")
	}
	TempDir = "/recordings"                            // recordings will end up here before they are converted
	StorageDir = "/srv/cephfs/livestream/rec/TUM-Live" // recordings will end up here after they are converted
	LrzUser = os.Getenv("LrzUser")
	LrzMail = os.Getenv("LrzMail")
	LrzPhone = os.Getenv("LrzPhone")
	LrzSubDir = os.Getenv("LrzSubDir")
	LrzUploadUrl = os.Getenv("LrzUploadUrl")
	MainBase = os.Getenv("MainBase") // eg. live.mm.rbg.tum.de
	LogDir = os.Getenv("LogDir")
	if LogDir == "" {
		LogDir = "/var/log/stream"
	}
	err := os.MkdirAll(LogDir, 0755)
	if err != nil {
		log.Warn("Could not create log directory: ", err)
	}

	// the hostname is required to announce this worker to the main server
	// Usually this is passed as an environment variable using docker. Otherwise, it is set to the hostname of the machine
	Hostname = os.Getenv("Host")
	if Hostname == "" {
		Hostname, err = os.Hostname()
		if err != nil {
			log.Fatalf("Could not get hostname: %v\n", err)
		}
	}

	// join main server:
	var conn *grpc.ClientConn
	// retry connecting to server every 5 seconds until successful
	for {
		conn, err = grpc.Dial(fmt.Sprintf("%s:50052", MainBase), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		} else {
			log.Warnf("Could not connect to main server: %v\n", err)
			time.Sleep(time.Second * 5)
		}
	}

	client := pb.NewFromWorkerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := client.JoinWorkers(ctx, &pb.JoinWorkersRequest{
		Token:    Token,
		Hostname: Hostname,
	})
	if err != nil {
		log.Fatalf("Could not join main server: %v\n", err)
	}
	WorkerID = resp.WorkerId
	log.Infof("Joined main server with worker id: %s\n", WorkerID)
}
