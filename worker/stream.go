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

	// in case ffmpeg dies retry until stream should be done.
	for time.Now().Before(streamUntil) {
		cmd := exec.Command(
			"ffmpeg", "-nostats", "-rtsp_transport", "tcp",
			"-t", fmt.Sprintf("%.0f", streamEnd.Sub(time.Now()).Seconds()), // timeout ffmpeg when stream is finished
			"-i", fmt.Sprintf("rtsp://%s", source),
			"-map", "0", "-c", "copy", "-f", "mpegts", "-", "-c:v", "libx264", "-preset", "veryfast", "-maxrate", "1500k", "-bufsize", "3000k", "-g", "50", "-r", "25", "-c:a", "aac", "-ar", "44100", "-b:a", "128k",
			"-f", "flv", fmt.Sprintf("%s%s", cfg.IngestBase, fileName))
		outfile, err := os.OpenFile(fmt.Sprintf("%s/%s.ts", cfg.TempDir, fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.WithError(err).Error("Unable to create file for recording")
			time.Sleep(time.Second) // sleep a second to prevent high load
			continue
		}
		cmd.Stdout = outfile
		err = cmd.Wait()
		if err != nil {
			log.WithError(err).Error("Error while streaming")
			continue
		}
	}
}
