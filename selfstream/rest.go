//Package selfstream handles notifications for self streaming from nginx
package selfstream

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/worker"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

//streams contains a map from streaming keys to their ids
var streams = make(map[string]*worker.StreamContext)

//InitApi creates routes for the api consumed by nginx
func InitApi(addr string) {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		if cfg.WorkerID == "" {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.String(http.StatusOK, "Hi, I'm alive, give me some work!")
	})
	r.POST("/on_publish", onPublish)
	r.POST("/on_publish_done", onPublishDone)
	r.POST("/on_record_done", onRecordDone)
	err := r.Run(addr)
	if err != nil {
		log.WithError(err).Fatal("Can't initialise self-streaming endpoints")
	}
}

//mustGetStreamInfo gets the stream key and slug from nginx requests and aborts with bad request if something is wrong
func mustGetStreamInfo(c *gin.Context) (streamKey string, slug string, err error) {
	name, e := c.GetPostForm("name")
	if !e {
		c.AbortWithStatus(http.StatusBadRequest)
		return "", "", errors.New("no stream slug")
	}
	tcUrl, e := c.GetPostForm("tcurl")
	if !e {
		c.AbortWithStatus(http.StatusBadRequest)
		return "", "", errors.New("no stream key")
	}
	if m, _ := regexp.MatchString(".+\\?secret=.+", tcUrl); !m {
		c.AbortWithStatus(http.StatusForbidden)
		return "", "", errors.New("stream key invalid")
	}
	key := strings.Split(tcUrl, "?secret=")[1]
	return key, name, nil
}
