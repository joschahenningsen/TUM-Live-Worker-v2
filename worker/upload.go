package worker

import (
	"bytes"
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker-v2/cfg"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func upload(streamCtx *StreamContext) {
	f, err := os.Open(streamCtx.getTranscodingFileName())
	if err != nil {
		log.Printf("unable to open converted file for upload: %v", err)
		return
	}
	values := map[string]io.Reader{
		"filename":    f,
		"benutzer":    strings.NewReader(cfg.LrzUser),
		"mailadresse": strings.NewReader(cfg.LrzMail),
		"telefon":     strings.NewReader(cfg.LrzPhone),
		"unidir":      strings.NewReader("tum"),
		"subdir":      strings.NewReader(cfg.LrzSubDir),
		"info":        strings.NewReader(""),
	}
	err = PostFileUpload(cfg.LrzUploadUrl, values)
	if err != nil {
		log.WithError(err).Error("Can't upload video to lrz")
		return
	}
}

// PostFileUpload - example kindly provided by Attila O. Thanks buddy!
func PostFileUpload(url string, values map[string]io.Reader) (err error) {
	client := http.DefaultClient
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}
