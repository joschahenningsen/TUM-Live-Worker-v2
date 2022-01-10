package worker

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

func transcode(streamCtx *StreamContext) {
	prepare(streamCtx.getTranscodingFileName())
	var cmd *exec.Cmd
	// create command fitting its content with appropriate niceness:
	in := streamCtx.getRecordingFileName()
	out := streamCtx.getTranscodingFileName()
	switch streamCtx.streamVersion {
	case "CAM":
		// compress camera image slightly more
		cmd = exec.Command("nice", "-n", "10", "ffmpeg", "-y", "-nostats", "-i", in, "-vsync", "2", "-c:v", "libx264", "-c:a", "aac", "-b:a", "128k", "-crf", "26", out)
	case "PRES":
		cmd = exec.Command("nice", "-n", "9", "ffmpeg", "-y", "-nostats", "-i", in, "-vsync", "2", "-c:v", "libx264", "-tune", "stillimage", "-c:a", "aac", "-b:a", "128k", "-crf", "20", out)
	case "COMB":
		cmd = exec.Command("nice", "-n", "8", "ffmpeg", "-y", "-nostats", "-i", in, "-vsync", "2", "-c:v", "libx264", "-c:a", "aac", "-b:a", "128k", "-crf", "24", out)
	default:
		//unknown source, use higher compression and less priority
		cmd = exec.Command("nice", "-n", "10", "ffmpeg", "-y", "-nostats", "-i", in, "-vsync", "2", "-c:v", "libx264", "-c:a", "aac", "-b:a", "128k", "-crf", "26", out)
	}
	log.WithFields(log.Fields{"input": in, "output": out, "command": cmd.String()}).Info("Transcoding")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"output": string(output)}).WithError(err).Error("Failed to process stream")
	} else {
		log.WithField("stream", streamCtx.getStreamName()).Info("Transcoding finished")
	}
	log.Info("Start Probing duration")
	duration, err := getDuration(streamCtx.getTranscodingFileName())
	if err != nil {
		log.WithError(err).Error("Failed to probe duration")
	} else {
		streamCtx.duration = uint32(duration)
		log.WithField("duration", duration).Info("Probing duration finished")
	}
}

// creates folder for output file if it doesn't exist
func prepare(out string) {
	dir := filepath.Dir(out)
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		log.WithError(err).Error("Could not create target folder for transcoding")
	}
}
