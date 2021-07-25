package worker

import (
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"testing"
	"time"
)

var s StreamContext

func setup() {
	cfg.WorkerID = "123"
	s = StreamContext{
		courseSlug:    "eidi",
		teachingTerm:  "W",
		teachingYear:  2021,
		startTime:     time.Date(2021, 9, 23, 8, 0, 0, 0, time.Local),
		streamId:      1,
		streamVersion: "COMB",
		publishVoD:    true,
		stream:        true,
		endTime:       time.Now().Add(time.Hour),
		commands:      nil,
	}
	cfg.TempDir = "/recordings"
}

func TestGetTranscodingFileName(t *testing.T) {
	setup()
	transcodingNameShould := "/srv/cephfs/livestream/rec/TUM-Live/2021/W/eidi/2021-09-23_08-00/eidi_2021-09-23_08-00_COMB.mp4"
	if got := s.getTranscodingFileName(); got != transcodingNameShould {
		t.Errorf("Wrong transcoding name, should be %s but is %s", transcodingNameShould, got)
	}
}

func TestGetRecordingFileName(t *testing.T) {
	setup()
	recordingNameShould := "/recordings/eidi_2021-09-23_08-00_COMB.ts"
	if got := s.getRecordingFileName(); got != recordingNameShould {
		t.Errorf("Wrong recording name, should be %s but is %s", recordingNameShould, got)
	}
}

