package worker

import (
	"context"
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

func closeConnection(conn *grpc.ClientConn) {
	err := conn.Close()
	if err != nil {
		log.WithError(err).Error("Could not close connection")
	}
}

func notifySilenceResults(silences *[]silence, streamID uint32) {
	if silences == nil {
		return
	}
	client, conn, err := GetClient()
	if err != nil {
		log.WithError(err).Error("Unable to dial server")
		return
	}
	defer closeConnection(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var starts []uint32
	var ends []uint32
	sArr := *silences
	for _, i := range sArr {
		starts = append(starts, uint32(i.Start))
		ends = append(ends, uint32(i.End))
	}

	_, err = client.NotifySilenceResults(ctx, &pb.SilenceResults{
		WorkerID: cfg.WorkerID,
		StreamID: streamID,
		Starts:   starts,
		Ends:     ends,
	})
	if err != nil {
		log.WithError(err).Error("Could not send silence replies")
	}
}

func notifyStreamStart(streamCtx *StreamContext) {
	client, conn, err := GetClient()
	if err != nil {
		log.WithError(err).Error("Unable to dial server")
		return
	}
	defer closeConnection(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := client.NotifyStreamStarted(ctx, &pb.StreamStarted{
		WorkerID:   cfg.WorkerID,
		StreamID:   streamCtx.streamId,
		HlsUrl:     fmt.Sprintf(streamCtx.outUrl, streamCtx.streamName), // could look like: fmt.Sprintf("https://live.stream.lrz.de/livetum/smil:%s_all.smil/playlist.m3u8?dvr", streamCtx.streamName)
		SourceType: streamCtx.streamVersion,
	})
	if err != nil || !resp.Ok {
		log.WithError(err).Error("Could not notify stream started")
	}
}

func notifyStreamDone(streamID uint32, endedEarly bool) {
	client, conn, err := GetClient()
	if err != nil {
		log.WithError(err).Error("Unable to dial server")
		return
	}
	defer closeConnection(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := client.NotifyStreamFinished(ctx, &pb.StreamFinished{
		WorkerID:   cfg.WorkerID,
		StreamID:   streamID,
		EndedEarly: endedEarly,
	})
	if err != nil || !resp.Ok {
		log.WithError(err).Error("Could not notify stream finished")
	}
}

func notifyTranscodingDone(streamCtx *StreamContext) {
	client, conn, err := GetClient()
	if err != nil {
		log.WithError(err).Error("Unable to dial server")
		return
	}
	defer closeConnection(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := client.NotifyTranscodingFinished(ctx, &pb.TranscodingFinished{
		WorkerID: cfg.WorkerID,
		StreamID: streamCtx.streamId,
		FilePath: streamCtx.getTranscodingFileName(),
		Duration: streamCtx.duration,
	})
	if err != nil || !resp.Ok {
		log.WithError(err).Error("Could not notify stream finished")
	}
}

func notifyUploadDone(streamCtx *StreamContext) {
	client, conn, err := GetClient()
	if err != nil {
		log.WithError(err).Error("Unable to dial server")
		return
	}
	defer closeConnection(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := client.NotifyUploadFinished(ctx, &pb.UploadFinished{
		WorkerID:   cfg.WorkerID,
		StreamID:   streamCtx.streamId,
		HLSUrl:     fmt.Sprintf("https://stream.lrz.de/vod/_definst_/mp4:tum/RBG/%s.mp4/playlist.m3u8", streamCtx.getStreamNameVoD()),
		SourceType: streamCtx.streamVersion,
	})
	if err != nil || !resp.Ok {
		log.WithError(err).Error("Could not notify upload finished")
	}
}

func GetClient() (pb.FromWorkerClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:50052", cfg.MainBase), grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	return pb.NewFromWorkerClient(conn), conn, nil
}
