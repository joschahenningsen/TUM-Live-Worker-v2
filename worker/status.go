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
var S Status

func init() {
	S = Status{
		workload: 0,
		Jobs:     []string{},
	}
	c := cron.New()
	_, err := c.AddFunc("* * * * *", S.SendHeartbeat)
	if err != nil {
		log.WithError(err).Error(":(")
	}
	c.Start()
}

const (
	costStream      = 2
	costTranscoding = 1
)

type Status struct {
	workload uint
	Jobs     []string
}

func (s Status) startStream(streamCtx *streamContext) {
	statusLock.Lock()
	// notify about stream:
	if server, err := grpc.Dial(fmt.Sprintf("%s:50052", cfg.MainBase), grpc.WithInsecure()); err != nil { // workerId used for authentication
		log.WithError(err).Error("unable to dial for stream start message")
		return
	} else {
		client := pb.NewFromWorkerClient(server)
		_, err := client.NotifyStreamStarted(context.Background(), &pb.StreamStarted{
			WorkerID:   cfg.WorkerID,
			StreamID:   streamCtx.streamId,
			HlsUrl:     fmt.Sprintf("https://live.stream.lrz.de/livetum/%s{{quality}}?dvr/playlist.m3u8", streamCtx.getStreamName()),
			SourceType: streamCtx.streamVersion,
		})
		if err != nil {
			log.WithError(err).Error("Sending Stream start event failed")
		}
	}
	defer statusLock.Unlock()
	s.workload += costStream
	s.Jobs = append(s.Jobs, fmt.Sprintf("streaming %s", streamCtx.getStreamName()))
}

func (s Status) startRecording(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	s.workload += costStream
	s.Jobs = append(s.Jobs, fmt.Sprintf("recording %s", name))
}

func (s Status) startTranscoding(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	s.workload += costTranscoding
	s.Jobs = append(s.Jobs, fmt.Sprintf("transcoding %s", name))
}

func (s Status) endStream(streamCtx *streamContext) {
	statusLock.Lock()
	defer statusLock.Unlock()
	s.workload -= costStream
	for i := range s.Jobs {
		if s.Jobs[i] == fmt.Sprintf("streaming %s", streamCtx.getStreamName()) {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			return
		}
	}
}

func (s Status) endRecording(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	s.workload -= costStream
	for i := range s.Jobs {
		if s.Jobs[i] == fmt.Sprintf("recording %s", name) {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			return
		}
	}
}

func (s Status) endTranscoding(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	s.workload -= costTranscoding
	for i := range s.Jobs {
		if s.Jobs[i] == fmt.Sprintf("transcoding %s", name) {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			return
		}
	}
}

func (s Status) SendHeartbeat() {
	log.Debug("sending HeartBeat")
	if server, err := grpc.Dial(fmt.Sprintf("%s:50052", cfg.MainBase), grpc.WithInsecure()); err != nil { // workerId used for authentication
		log.WithError(err).Error("unable to dial for heartbeat")
		return
	} else {
		client := pb.NewFromWorkerClient(server)
		_, err := client.SendHeartBeat(context.Background(), &pb.HeartBeat{
			WorkerID: cfg.WorkerID,
			Workload: uint32(s.workload),
			Jobs:     s.Jobs,
		})
		if err != nil {
			log.WithError(err).Error("Sending Heartbeat failed")
			return
		}
	}
}
