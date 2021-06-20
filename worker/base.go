package worker

import (
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	log "github.com/sirupsen/logrus"
)

func HandleStreamRequest(request *pb.StreamRequest) {
	log.WithField("request", request).Info("Request to stream")

	//e.g. eidi_2006_01_02-15:04_PRES
	fileName := fmt.Sprintf("%s_%s_%s", request.CourseSlug, request.Start.AsTime().Format("2006_01_02-15:04"), request.SourceType)
	status.workload += 2
	if !request.PublishStream {
		status.startRecording(fileName)
		log.Info("only recording")
		record(request.SourceUrl, request.End.AsTime(), fileName)
		status.endRecording(fileName)
	} else {
		status.startStream(fileName)
		log.Info("record and stream")
		stream(request.SourceUrl, request.End.AsTime(), fileName)
		status.endStream(fileName)
	}
	transcode(request.SourceType, fmt.Sprintf("%s/%s.ts", cfg.TempDir, fileName), "")
	// todo: check health of output file and delete temp
	if request.PublishVoD {
		upload("")
	}
}
