package worker

import (
	"fmt"
	"sync"
)

var status = Status{
	workload: 0,
	Jobs:     []string{},
}
var statusLock = sync.RWMutex{}

const (
	costStream      = 2
	costTranscoding = 1
)

type Status struct {
	workload uint
	Jobs     []string
}

func (s Status) startStream(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	status.workload += costStream
	status.Jobs = append(status.Jobs, fmt.Sprintf("streaming %s", name))
}

func (s Status) startRecording(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	status.workload += costStream
	status.Jobs = append(status.Jobs, fmt.Sprintf("recording %s", name))
}

func (s Status) startTranscoding(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	status.workload += costTranscoding
	status.Jobs = append(status.Jobs, fmt.Sprintf("transcoding %s", name))
}

func (s Status) endStream(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	status.workload -= costStream
	for i := range status.Jobs {
		if status.Jobs[i] == fmt.Sprintf("streaming %s", name) {
			status.Jobs = append(status.Jobs[:i], status.Jobs[i+1:]...)
			return
		}
	}
}

func (s Status) endRecording(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	status.workload -= costStream
	for i := range status.Jobs {
		if status.Jobs[i] == fmt.Sprintf("recording %s", name) {
			status.Jobs = append(status.Jobs[:i], status.Jobs[i+1:]...)
			return
		}
	}
}

func (s Status) endTranscoding(name string) {
	statusLock.Lock()
	defer statusLock.Unlock()
	status.workload -= costTranscoding
	for i := range status.Jobs {
		if status.Jobs[i] == fmt.Sprintf("transcoding %s", name) {
			status.Jobs = append(status.Jobs[:i], status.Jobs[i+1:]...)
			return
		}
	}
}
