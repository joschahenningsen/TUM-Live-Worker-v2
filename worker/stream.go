package worker

import (
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"time"
)

//stream records and streams a lecture hall to the lrz
func stream(streamCtx *StreamContext) {
	// add 10 minutes padding to stream end in case lecturers do lecturer things
	streamUntil := streamCtx.endTime.Add(time.Minute * 10)
	log.WithFields(log.Fields{"source": streamCtx.sourceUrl, "end": streamUntil, "fileName": streamCtx.getRecordingFileName()}).
		Info("streaming lecture hall")
	S.startStream(streamCtx)
	defer S.endStream(streamCtx)
	// in case ffmpeg dies retry until stream should be done.
	lastErr := time.Now().Add(time.Minute * -1)
	for time.Now().Before(streamUntil) && !streamCtx.stopped {
		var cmd *exec.Cmd

		// todo: too much duplication
		if strings.Contains(streamCtx.sourceUrl, "rtsp") {
			cmd = exec.Command(
				"bash", "-c",
				`ffmpeg -hide_banner -nostats -rtsp_transport tcp -stimeout 5000000 -t ` + fmt.Sprintf("%.0f", streamUntil.Sub(time.Now()).Seconds()) + // timeout ffmpeg when stream is finished
				" -i " + fmt.Sprintf(streamCtx.sourceUrl) +
				` -map 0 -c copy -f mpegts - -c:v libx264 -preset veryfast -tune zerolatency -maxrate 2500k -bufsize 3000k -g 60 -r 30 -x264-params keyint=60:scenecut=0 -c:a aac -ar 44100 -b:a 128k `+
				`-f flv ` + fmt.Sprintf("%s/%s", streamCtx.ingestServer, streamCtx.streamName) + " >> " + streamCtx.getRecordingFileName())
		} else {
			cmd = exec.Command(
				"bash", "-c",
				`ffmpeg -hide_banner -nostats -stimeout 5000000 -t ` + fmt.Sprintf("%.0f", streamUntil.Sub(time.Now()).Seconds()) + // timeout ffmpeg when stream is finished
					" -i " + fmt.Sprintf(streamCtx.sourceUrl) +
					` -map 0 -c copy -f mpegts - -c:v libx264 -preset veryfast -tune zerolatency -maxrate 2500k -bufsize 3000k -g 60 -r 30 -x264-params keyint=60:scenecut=0 -c:a aac -ar 44100 -b:a 128k `+
					`-f flv ` + fmt.Sprintf("%s/%s", streamCtx.ingestServer, streamCtx.streamName) + " >> " + streamCtx.getRecordingFileName())
		}
		// persist stream command in context, so it can be killed later
		streamCtx.streamCmd = cmd

		log.WithField("cmd", cmd.String()).Info("Starting stream")
		outfile, err := os.OpenFile(streamCtx.getRecordingFileName(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			errorWithBackoff(&lastErr, "Unable to create file for recording", err)
			continue
		}
		cmd.Stdout = outfile
		ffmpegErr, errFfmpegErrFile := os.OpenFile(fmt.Sprintf("%s/ffmpeg_%s.log", cfg.LogDir, streamCtx.getStreamName()), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if errFfmpegErrFile == nil {
			cmd.Stderr = ffmpegErr
		} else {
			log.WithError(errFfmpegErrFile).Error("Could not create file for ffmpeg stdErr")
		}
		err = cmd.Run()
		if err != nil && !streamCtx.stopped {
			errorWithBackoff(&lastErr, "Error while streaming (run)", err)
			if errFfmpegErrFile == nil {
				_ = ffmpegErr.Close()
			}
			_ = outfile.Close()
			continue
		}
		if errFfmpegErrFile == nil {
			_ = ffmpegErr.Close()
		}
		_ = outfile.Close()
	}
}

//errorWithBackoff updates lastError and sleeps for a second if the last error was within this second
func errorWithBackoff(lastError *time.Time, msg string, err error) {
	log.WithFields(log.Fields{"lastErr": lastError}).WithError(err).Error(msg)
	if time.Now().Add(time.Second * -1).Before(*lastError) {
		log.Warn("too many errors, backing off a second.")
		time.Sleep(time.Second)
	}
	now := time.Now()
	*lastError = now
}
