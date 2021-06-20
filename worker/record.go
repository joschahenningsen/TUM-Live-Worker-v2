package worker

import (
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

//record records a source until endTime +10 minutes without pushing to lrz
func record(source string, recordEnd time.Time, fileName string) {
	// add 10 minutes padding to stream end in case lecturers do lecturer things
	recordUntil := recordEnd.Add(time.Minute * 10)
	log.WithFields(log.Fields{"source": source, "end": recordUntil, "fileName": fileName}).
		Info("Recording lecture hall")

	// in case ffmpeg dies retry until stream should be done.
	for time.Now().Before(recordUntil) {
		cmd := exec.Command(
			"ffmpeg", "-nostats", "-rtsp_transport", "tcp",
			"-t", fmt.Sprintf("%.0f", recordUntil.Sub(time.Now()).Seconds()), // timeout ffmpeg when stream is finished
			"-i", fmt.Sprintf("rtsp://%s", source),
			"-map", "0",
			"-c:v", "copy",
			"-c:a", "copy",
			"-f", "mpegts", "-")
		outfile, err := os.OpenFile(fmt.Sprintf("%s/%s.ts", cfg.TempDir, fileName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.WithError(err).Error("Unable to create file for recording")
			time.Sleep(time.Second) // sleep a second to prevent high load
			continue
		}
		cmd.Stdout = outfile
		err = cmd.Wait()
		if err != nil {
			log.WithError(err).Error("Error while recording")
			continue
		}
	}
}
