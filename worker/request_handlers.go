package worker

import (
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	log "github.com/sirupsen/logrus"
	"time"
)

func HandleStreamRequest(request *pb.StreamRequest) {
	log.WithField("request", request).Info("Request to stream")

	//setup context with relevant information to pass to other subprocesses
	streamCtx := &streamContext{
		streamId:      request.GetStreamID(),
		sourceUrl:     request.GetSourceUrl(),
		courseSlug:    request.GetCourseSlug(),
		teachingTerm:  request.GetCourseTerm(),
		teachingYear:  request.GetCourseYear(),
		startTime:     request.GetStart().AsTime(),
		endTime:       request.GetEnd().AsTime(),
		streamVersion: request.GetSourceType(),
		publishVoD:    request.GetPublishVoD(),
		stream:        request.GetPublishStream(),
	}

	//only record
	if !streamCtx.stream {
		S.startRecording(streamCtx.getRecordingFileName())
		log.Info("only recording")
		record(streamCtx)
		S.endRecording(streamCtx.getRecordingFileName())
	} else {
		log.Info("record and stream")
		stream(streamCtx)
	}
	// notify stream/recording done
	notifyStreamDone(streamCtx.streamId)

	transcode(streamCtx)
	notifyTranscodingDone(streamCtx)
	// todo: check health of output file and delete temp
	if request.PublishVoD {
		upload(streamCtx)
		notifyUploadDone(streamCtx)
	}
}

// streamContext contains all important information on a stream
type streamContext struct {
	streamId      uint32         //id of the stream
	sourceUrl     string         //url of the streams source, e.g. 10.0.0.4
	courseSlug    string         //slug of the course, e.g. eidi
	teachingTerm  string         //S or W depending on the courses teaching-term
	teachingYear  uint32         //Year the course takes place in
	startTime     time.Time      //time the stream should start
	endTime       time.Time      //end of the stream (including +10 minute safety)
	streamVersion string         //version of the stream to be handled, e.g. PRES, COMB or CAM
	publishVoD    bool           //whether file should be uploaded
	stream        bool           //whether streaming is enabled
	commands      map[string]int //map command type to pid, e.g. "stream"->123
}

// getRecordingFileName returns the filename a stream should be saved to before transcoding.
// example: /recordings/eidi_2021-09-23_10-00_COMB.ts
func (s streamContext) getRecordingFileName() string {
	return fmt.Sprintf("%s/%s.ts",
		cfg.TempDir,
		s.getStreamName())
}

// getTranscodingFileName returns the filename a stream should be saved to after transcoding.
// example: /srv/sharedMassStorage/2021/S/eidi/2021-09-23_10-00/eidi_2021-09-23_10-00_PRES.mp4
func (s streamContext) getTranscodingFileName() string {
	return fmt.Sprintf("%s/%d/%s/%s/%s/%s.mp4",
		cfg.StorageDir,
		s.teachingYear,
		s.teachingTerm,
		s.courseSlug,
		s.startTime.Format("2006-01-02_15-04"),
		s.getStreamName())
}

func (s streamContext) getStreamName() string {
	return fmt.Sprintf("%s-%s-%s",
		s.courseSlug,
		s.startTime.Format("2006-01-02-15-04"),
		s.streamVersion)
}
