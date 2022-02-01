//Package selfstream handles notifications for self streaming from nginx
package selfstream

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/worker"
	log "github.com/sirupsen/logrus"
)

// streams contains a map from streaming keys to their ids
var streams = safeStreams{streams: make(map[string]*worker.StreamContext)}

type safeStreams struct {
	mutex   sync.Mutex
	streams map[string]*worker.StreamContext
}

// InitApi creates routes for the api consumed by nginx
func InitApi(addr string) {
	defaultHandler := func(w http.ResponseWriter, _ *http.Request) {
		if cfg.WorkerID == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, "Hi, I'm alive, give me some work!\n")
	}
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/on_publish", streams.onPublish)
	http.HandleFunc("/on_publish_done", streams.onPublishDone)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// mustGetStreamInfo gets the stream key and slug from nginx requests and aborts with bad request if something is wrong
func mustGetStreamInfo(w http.ResponseWriter, r *http.Request) (streamKey string, slug string, err error) {
	name := r.Form.Get("name")
	if name == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return "", "", errors.New("no stream slug")
	}
	tcUrl := r.Form.Get("tcurl")
	if tcUrl == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return "", "", errors.New("no stream key")
	}
	if m, _ := regexp.MatchString(".+\\?secret=.+", tcUrl); !m {
		w.WriteHeader(http.StatusForbidden)
		return "", "", errors.New("stream key invalid")
	}
	key := strings.Split(tcUrl, "?secret=")[1]
	return key, name, nil
}
