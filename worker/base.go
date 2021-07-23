package worker

import (
	"context"
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func HandleStreamRequest(request *pb.StreamRequest) {
	log.WithField("request", request).Info("Request to stream")

	//e.g. eidi_2006_01_02-15:04_PRES
	fileName := fmt.Sprintf("%s_%s_%s", request.CourseSlug, request.Start.AsTime().Format("2006_01_02-15:04"), request.SourceType)

	if !request.PublishStream {
		S.startRecording(fileName)
		log.Info("only recording")
		record(request.SourceUrl, request.End.AsTime(), fileName)
		S.endRecording(fileName)
	} else {
		S.startStream(fileName)
		log.Info("record and stream")
		stream(request.SourceUrl, request.End.AsTime(), fileName)
		S.endStream(fileName)
	}
	// notify stream/recording done
	conn, err := grpc.Dial(fmt.Sprintf("%s:50052", cfg.MainBase), grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Error("Unable to dial server")
		return
	}

	client := pb.NewFromWorkerClient(conn)
	resp, err := client.NotifyStreamFinished(context.Background(), &pb.StreamFinished{
		WorkerID:   cfg.WorkerID,
		StreamID:   request.StreamID,
	})
	if err != nil || !resp.Ok {
		log.WithError(err).Error("Could not notify stream finished")
	}
	transcode(request.SourceType, fmt.Sprintf("%s/%s.ts", cfg.TempDir, fileName), "")
	// todo: check health of output file and delete temp
	if request.PublishVoD {
		upload("")
	}
	resp, err = client.NotifyTranscodingFinished(context.Background(), &pb.TranscodingFinished{
		WorkerID:   cfg.WorkerID,
		StreamID:   request.StreamID,
		FilePath:   fileName,
		HlsUrl:     fmt.Sprintf("https://live.stream.lrz.de/livetum/%s/playlist.m3u8", fileName),
		SourceType: request.SourceType,
	})
	if err != nil || !resp.Ok {
		log.WithError(err).Error("Could not notify transcoding finished")
		return
	}
	// notify transcoding done:
	_ = conn.Close()
}
