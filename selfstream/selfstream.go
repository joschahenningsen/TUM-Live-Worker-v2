package selfstream

import (
	"github.com/gin-gonic/gin"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/worker"
	log "github.com/sirupsen/logrus"
	"net/http"
)

//deprecated
//onRecordDone is called by nginx when the recording is finished
func onRecordDone(c *gin.Context) {
	log.Info("On record done called")
	streamKey, _, err := mustGetStreamInfo(c)
	if err != nil {
		log.WithFields(log.Fields{"request": c.Request.Form}).WithError(err).Warn("onRecordDone: bad on_publish request")
		return
	}
	if streamCtx, ok := streams[streamKey]; ok {
		worker.HandleSelfStreamRecordEnd(streamCtx)
	} else {
		log.WithField("streamKey", streamKey).Error("stream key not existing in self streams map")
	}
}

//onPublishDone is called by nginx when the stream stops publishing
func onPublishDone(c *gin.Context) {
	log.Info("On publish done called")
	streamKey, _, err := mustGetStreamInfo(c)
	if err != nil {
		log.WithFields(log.Fields{"request": c.Request.Form}).WithError(err).Warn("onRecordDone: bad on_publish request")
		return
	}
	if streamCtx, ok := streams[streamKey]; ok {
		go func() {
			worker.HandleStreamEnd(streamCtx)
			worker.HandleSelfStreamRecordEnd(streamCtx)
		}()
	} else {
		log.WithField("streamKey", streamKey).Error("stream key not existing in self streams map")
	}
}

//onPublish is called by nginx when the stream starts publishing
func onPublish(c *gin.Context) {
	log.Info("On publish called")
	streamKey, slug, err := mustGetStreamInfo(c)
	if err != nil {
		log.WithFields(log.Fields{"request": c.Request.Form}).WithError(err).Warn("onRecordDone: bad on_publish request")
		return
	}
	client, conn, err := worker.GetClient()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	resp, err := client.SendSelfStreamRequest(c, &pb.SelfStreamRequest{
		WorkerID:   cfg.WorkerID,
		StreamKey:  streamKey,
		CourseSlug: slug,
	})
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		_ = conn.Close()
		return
	}
	// register stream in local map
	streamContext := worker.HandleSelfStream(resp, slug)
	streams[streamKey] = streamContext
	_ = conn.Close()
}
