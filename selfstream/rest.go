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
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusInternalServerError)
		}
		if cfg.WorkerID == "" {
			http.Error(w, "Worker has no ID", http.StatusInternalServerError)
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
		http.Error(w, "Did not find stream slug", http.StatusInternalServerError)
		return "", "", errors.New("no stream slug")
	}
	tcUrl := r.Form.Get("tcurl")
	if tcUrl == "" {
		http.Error(w, "Did not find stream key.", http.StatusInternalServerError)
		return "", "", errors.New("no stream key")
	}
	if m, _ := regexp.MatchString(".+\\?secret=.+", tcUrl); !m {
		http.Error(w, "Stream key in request is invalid", http.StatusInternalServerError)
		return "", "", errors.New("stream key invalid")
	}
	key := strings.Split(tcUrl, "?secret=")[1]
	return key, name, nil
}
