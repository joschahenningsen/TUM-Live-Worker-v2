package selfstream

import (
	"net/http"

	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/pb"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/worker"
	log "github.com/sirupsen/logrus"
)

// onPublishDone is called by nginx when the stream stops publishing
func (s *safeStreams) onPublishDone(w http.ResponseWriter, r *http.Request) {
	log.Info("On publish done called")
	streamKey, _, err := mustGetStreamInfo(w, r)
	if err != nil {
		log.WithFields(log.Fields{"request": r.Form}).WithError(err).Warn("onRecordDone: bad on_publish request")
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if streamCtx, ok := s.streams[streamKey]; ok {
		go func() {
			worker.HandleStreamEnd(streamCtx)
			worker.NotifyStreamDone(streamCtx)
			worker.HandleSelfStreamRecordEnd(streamCtx)
		}()
	} else {
		log.WithField("streamKey", streamKey).Error("stream key not existing in self streams map")
	}
}

//onPublish is called by nginx when the stream starts publishing
func (s *safeStreams) onPublish(w http.ResponseWriter, r *http.Request) {
	log.Info("On publish called")
	streamKey, slug, err := mustGetStreamInfo(w, r)
	if err != nil {
		log.WithFields(log.Fields{"request": r.Form}).WithError(err).Warn("onRecordDone: bad on_publish request")
		return
	}
	client, conn, err := worker.GetClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := client.SendSelfStreamRequest(r.Context(), &pb.SelfStreamRequest{
		WorkerID:   cfg.WorkerID,
		StreamKey:  streamKey,
		CourseSlug: slug,
	})
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		_ = conn.Close()
		return
	}
	// register stream in local map
	streamContext := worker.HandleSelfStream(resp, slug)

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.streams[streamKey] = streamContext
	_ = conn.Close()
}
