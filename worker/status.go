package worker

import (
	"context"
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"sync"
)

var statusLock = sync.RWMutex{}
var S *Status

func init() {
	S = &Status{
		workload: 0,
		Jobs:     []string{},
	}
	c := cron.New()
	_, _ = c.AddFunc("* * * * *", S.SendHeartbeat)
	c.Start()
}

const (
	costStream           = 3
	costTranscoding      = 2
	costSilenceDetection = 1
)

type Status struct {
	workload uint
	Jobs     []string
}

func (s *Status) startSilenceDetection(streamCtx *StreamContext) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	s.workload += costSilenceDetection
	s.Jobs = append(s.Jobs, fmt.Sprintf("detecting silence in %s", streamCtx.getStreamName()))
	statusLock.Unlock()
}

func (s *Status) startStream(streamCtx *StreamContext) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	notifyStreamStart(streamCtx)
	defer statusLock.Unlock()
	s.workload += costStream
	s.Jobs = append(s.Jobs, fmt.Sprintf("streaming %s", streamCtx.getStreamName()))
}

func (s *Status) startRecording(name string) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	defer statusLock.Unlock()
	s.workload += costStream
	s.Jobs = append(s.Jobs, fmt.Sprintf("recording %s", name))
}

func (s *Status) startTranscoding(name string) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	defer statusLock.Unlock()
	s.workload += costTranscoding
	s.Jobs = append(s.Jobs, fmt.Sprintf("transcoding %s", name))
}

func (s *Status) endStream(streamCtx *StreamContext) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	s.workload -= costStream
	for i := range s.Jobs {
		if s.Jobs[i] == fmt.Sprintf("streaming %s", streamCtx.getStreamName()) {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			break
		}
	}
	statusLock.Unlock()
}

func (s *Status) endRecording(name string) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	s.workload -= costStream
	for i := range s.Jobs {
		if s.Jobs[i] == fmt.Sprintf("recording %s", name) {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			break
		}
	}
	statusLock.Unlock()
}

func (s *Status) endTranscoding(name string) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	s.workload -= costTranscoding
	for i := range s.Jobs {
		if s.Jobs[i] == fmt.Sprintf("transcoding %s", name) {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			break
		}
	}
	statusLock.Unlock()
}

func (s *Status) endSilenceDetection(streamCtx *StreamContext) {
	defer s.SendHeartbeat()
	statusLock.Lock()
	s.workload -= s.workload
	for i := range s.Jobs {
		if s.Jobs[i] == fmt.Sprintf("detecting silence in %s", streamCtx.getStreamName()) {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			break
		}
	}
	statusLock.Unlock()
}

func (s *Status) SendHeartbeat() {
	// WithInsecure: workerId used for authentication, all servers are inside their own VLAN to further improve security
	if server, err := grpc.Dial(fmt.Sprintf("%s:50052", cfg.MainBase), grpc.WithInsecure()); err != nil {
		log.WithError(err).Error("unable to dial for heartbeat")
	} else {
		client := pb.NewFromWorkerClient(server)
		_, err := client.SendHeartBeat(context.Background(), &pb.HeartBeat{
			WorkerID: cfg.WorkerID,
			Workload: uint32(s.workload),
			Jobs:     s.Jobs,
		})
		if err != nil {
			log.WithError(err).Error("Sending Heartbeat failed")
		}
	}
}
