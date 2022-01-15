package worker

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	log "github.com/sirupsen/logrus"
)

// lectureHallStreams contains a map from streaming keys to their contexts.
// Note that we can have multiple contexts for different sources.
var lectureHallStreams = make(map[uint32][]*StreamContext)

func remove(contexts []*StreamContext, context *StreamContext) []*StreamContext {
	var newContexts []*StreamContext

	for _, i := range contexts {
		if i.sourceUrl != context.sourceUrl {
			newContexts = append(newContexts, i)
		}
	}
	log.Info("Removed stream with source: ", context.sourceUrl)
	return newContexts
}

func HandlePremiere(request *pb.PremiereRequest) {
	streamCtx := &StreamContext{
		streamId:      request.StreamID,
		sourceUrl:     request.FilePath,
		startTime:     time.Now(),
		streamVersion: "",
		courseSlug:    "PREMIERE",
		stream:        true,
		commands:      nil,
		ingestServer:  request.IngestServer,
		outUrl:        request.OutUrl,
	}
	// Register worker for premiere
	if !streamCtx.isSelfStream {
		lectureHallStreams[streamCtx.streamId] = append(lectureHallStreams[streamCtx.streamId], streamCtx)
	}
	S.startStream(streamCtx)
	streamPremiere(streamCtx)
	S.endStream(streamCtx)
	notifyStreamDone(streamCtx.streamId, streamCtx.endedEarly)
}

func HandleSelfStream(request *pb.SelfStreamResponse, slug string) *StreamContext {
	streamCtx := &StreamContext{
		streamId:      request.GetStreamID(),
		courseSlug:    request.GetCourseSlug(),
		teachingTerm:  request.GetCourseTerm(),
		teachingYear:  request.GetCourseYear(),
		startTime:     request.GetStreamStart().AsTime().Local(),
		endTime:       time.Now().Add(time.Hour * 7),
		publishVoD:    request.GetUploadVoD(),
		stream:        true,
		streamVersion: "COMB",
		isSelfStream:  false,
		ingestServer:  request.IngestServer,
		sourceUrl:     "rtmp://localhost/stream/" + slug,
		streamName:    request.StreamName,
		outUrl:        request.OutUrl,
	}
	go stream(streamCtx)
	return streamCtx
}

func HandleSelfStreamRecordEnd(ctx *StreamContext) {
	S.startTranscoding(ctx.getStreamName())
	transcode(ctx)
	S.endTranscoding(ctx.getStreamName())
	notifyTranscodingDone(ctx)
	if ctx.publishVoD {
		upload(ctx)
		notifyUploadDone(ctx)
	}
	S.startSilenceDetection(ctx)
	defer S.endSilenceDetection(ctx)

	sd := NewSilenceDetector(ctx.getTranscodingFileName())
	err := sd.ParseSilence()
	if err != nil {
		log.WithField("File", ctx.getTranscodingFileName()).WithError(err).Error("Detecting silence failed.")
		return
	}
	notifySilenceResults(sd.Silences, ctx.streamId)
}

func HandleStreamEndRequest(request *pb.EndStreamRequest) {
	log.Info("Attempting to end stream: ", request.StreamID)
	streams := lectureHallStreams[request.StreamID]
	for _, stream := range streams {
		streamCtx := stream
		if streamCtx.isSelfStream {
			log.Error("Unexpected self stream end request.")
			continue
		}
		streamCtx.discardVoD = request.DiscardVoD
		streamCtx.endedEarly = true
		// Register worker for stream
		lectureHallStreams[streamCtx.streamId] = append(lectureHallStreams[streamCtx.streamId], streamCtx)
		HandleStreamEnd(streamCtx)
	}
}

//HandleStreamEnd stops the ffmpeg instance by sending a SIGINT to it and prevents the loop to restart it by marking the stream context as stopped.
func HandleStreamEnd(ctx *StreamContext) {
	ctx.stopped = true
	if ctx.streamCmd != nil && ctx.streamCmd.Process != nil {
		err := ctx.streamCmd.Process.Kill()
		if err != nil {
			log.WithError(err).Warn("can't kill ffmpeg process")
		}
	} else {
		log.Warn("context has no command or process to end")
	}
	S.endStream(ctx)
	// Self-Streams notify here
	if ctx.isSelfStream {
		notifyStreamDone(ctx.streamId, ctx.endedEarly)
	}
	lectureHallStreams[ctx.streamId] = remove(lectureHallStreams[ctx.streamId], ctx)
}

func HandleStreamRequest(request *pb.StreamRequest) {
	log.WithField("request", request).Info("Request to stream")
	//setup context with relevant information to pass to other subprocesses
	streamCtx := &StreamContext{
		streamId:      request.GetStreamID(),
		sourceUrl:     "rtsp://" + request.GetSourceUrl(),
		courseSlug:    request.GetCourseSlug(),
		teachingTerm:  request.GetCourseTerm(),
		teachingYear:  request.GetCourseYear(),
		startTime:     request.GetStart().AsTime().Local(),
		endTime:       request.GetEnd().AsTime().Local(),
		streamVersion: request.GetSourceType(),
		publishVoD:    request.GetPublishVoD(),
		stream:        request.GetPublishStream(),
		streamName:    request.GetStreamName(),
		ingestServer:  request.GetIngestServer(),
		isSelfStream:  false,
		outUrl:        request.GetOutUrl(),
	}

	// Register worker for stream
	if !streamCtx.isSelfStream {
		lectureHallStreams[streamCtx.streamId] = append(lectureHallStreams[streamCtx.streamId], streamCtx)
	}

	//only record
	if !streamCtx.stream {
		S.startRecording(streamCtx.getRecordingFileName())
		record(streamCtx)
		S.endRecording(streamCtx.getRecordingFileName())
	} else {
		stream(streamCtx)
	}
	// notify stream/recording done
	notifyStreamDone(streamCtx.streamId, streamCtx.endedEarly)
	if streamCtx.discardVoD {
		return
	}

	S.startTranscoding(streamCtx.getStreamName())
	transcode(streamCtx)
	S.endTranscoding(streamCtx.getStreamName())
	notifyTranscodingDone(streamCtx)
	// todo: check health of output file and delete temp
	if request.PublishVoD {
		upload(streamCtx)
		notifyUploadDone(streamCtx)
	}

	if streamCtx.streamVersion == "COMB" {
		S.startSilenceDetection(streamCtx)
		defer S.endSilenceDetection(streamCtx)
		sd := NewSilenceDetector(streamCtx.getTranscodingFileName())
		err := sd.ParseSilence()
		if err != nil {
			log.WithField("File", streamCtx.getTranscodingFileName()).WithError(err).Error("Detecting silence failed.")
			return
		}
		notifySilenceResults(sd.Silences, streamCtx.streamId)
	}
}

// StreamContext contains all important information on a stream
type StreamContext struct {
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
	streamCmd     *exec.Cmd      // command used for streaming
	isSelfStream  bool           //deprecated
	streamName    string         // ingest target
	ingestServer  string         // ingest server e.g. rtmp://user:password@my.server
	stopped       bool           // whether the stream has been stopped
	outUrl        string         // url the stream will be available at
	discardVoD    bool           // whether the VoD should be discarded
	endedEarly    bool           // whether a stream was ended before it's official end

	// calculated after stream:
	duration uint32 //duration of the stream in seconds
}

// getRecordingFileName returns the filename a stream should be saved to before transcoding.
// example: /recordings/eidi_2021-09-23_10-00_COMB.ts
func (s StreamContext) getRecordingFileName() string {
	if !s.isSelfStream {
		return fmt.Sprintf("%s/%s.ts",
			cfg.TempDir,
			s.getStreamName())
	}
	return fmt.Sprintf("%s/%s_%s.flv",
		cfg.TempDir,
		s.courseSlug,
		s.startTime.Format("02012006"))
}

// getTranscodingFileName returns the filename a stream should be saved to after transcoding.
// example: /srv/sharedMassStorage/2021/S/eidi/2021-09-23_10-00/eidi_2021-09-23_10-00_PRES.mp4
func (s StreamContext) getTranscodingFileName() string {
	if s.isSelfStream {
		return fmt.Sprintf("%s/%d/%s/%s/%s/%s-%s.mp4",
			cfg.StorageDir,
			s.teachingYear,
			s.teachingTerm,
			s.courseSlug,
			s.startTime.Format("2006-01-02_15-04"),
			s.courseSlug,
			s.startTime.Format("02012006"))
	}
	return fmt.Sprintf("%s/%d/%s/%s/%s/%s.mp4",
		cfg.StorageDir,
		s.teachingYear,
		s.teachingTerm,
		s.courseSlug,
		s.startTime.Format("2006-01-02_15-04"),
		s.getStreamName())
}

// getStreamName returns the stream name, used for the worker status
func (s StreamContext) getStreamName() string {
	if !s.isSelfStream {
		return fmt.Sprintf("%s-%s%s",
			s.courseSlug,
			s.startTime.Format("2006-01-02-15-04"),
			s.streamVersion)
	}
	return s.courseSlug
}

// getStreamNameVoD returns the stream name for vod (lrz replaces - with _)
func (s StreamContext) getStreamNameVoD() string {
	if !s.isSelfStream {
		return strings.ReplaceAll(fmt.Sprintf("%s_%s%s",
			s.courseSlug,
			s.startTime.Format("2006_01_02_15_04"),
			s.streamVersion), "-", "_")
	}
	return s.courseSlug
}
