package worker

import (
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

//stream records and streams a lecture hall to the lrz
func stream(source string, streamEnd time.Time, fileName string) {
	// add 10 minutes padding to stream end in case lecturers do lecturer things
	streamUntil := streamEnd.Add(time.Minute * 10)
	log.WithFields(log.Fields{"source": source, "end": streamUntil, "fileName": fileName}).
		Info("Recording lecture hall")
	S.startStream(fileName)
	defer S.endStream(fileName)
	// in case ffmpeg dies retry until stream should be done.
	lastErr := time.Now().Add(time.Minute * -1)
	for time.Now().Before(streamUntil) {
		cmd := exec.Command(
			"ffmpeg", "-hide_banner", "-nostats", "-rtsp_transport", "tcp",
			"-stimeout", "5000000", // timeout ffmpeg when there's no data from the device for 5 seconds
			"-t", fmt.Sprintf("%.0f", streamEnd.Sub(time.Now()).Seconds()), // timeout ffmpeg when stream is finished
			"-i", fmt.Sprintf("rtsp://%s", source),
			"-map", "0", "-c", "copy", "-f", "mpegts", "-", "-c:v", "libx264", "-preset", "veryfast", "-maxrate", "1500k", "-bufsize", "3000k", "-g", "50", "-r", "25", "-c:a", "aac", "-ar", "44100", "-b:a", "128k",
			"-f", "flv", fmt.Sprintf("%s%s", cfg.IngestBase, fileName))
		log.WithField("cmd", cmd.String()).Info("Starting stream")
		outfile, err := os.OpenFile(fmt.Sprintf("%s/%s.ts", cfg.TempDir, fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			errorWithBackoff(&lastErr, "Unable to create file for recording", err)
			continue
		}
		cmd.Stdout = outfile
		ffmpegErr, errFfmpegErrFile := os.OpenFile(fmt.Sprintf("/var/log/tumliveworker/ffmpeg_%s.log", fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if errFfmpegErrFile == nil {
			cmd.Stderr = ffmpegErr
		} else {
			log.WithError(errFfmpegErrFile).Error("Could not create file for ffmpeg stdErr")
		}
		err = cmd.Start()
		if err != nil {
			errorWithBackoff(&lastErr, "Error while streaming (start)", err)
			if errFfmpegErrFile == nil {
				_ = ffmpegErr.Close()
			}
			continue
		}
		err = cmd.Wait()
		if err != nil {
			errorWithBackoff(&lastErr, "Error while streaming (wait)", err)
			if errFfmpegErrFile == nil {
				_ = ffmpegErr.Close()
			}
			continue
		}
		if errFfmpegErrFile == nil {
			_ = ffmpegErr.Close()
		}
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
