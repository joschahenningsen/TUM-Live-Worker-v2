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
	"os/exec"
)

func upload(streamCtx *StreamContext) {
	log.WithField("stream", streamCtx.getStreamName()).Info("Uploading stream")
	err := post(streamCtx.getTranscodingFileName())
	if err != nil {
        log.WithField("stream", streamCtx.getStreamName()).WithError(err).Error("Error uploading stream")
    }
	log.WithField("stream", streamCtx.getStreamName()).Info("Uploaded stream")
}

// post a file via curl
func post(file string) error {
	cmd := exec.Command("curl", "-F",
		"file=@"+file,
		"-F", "benutzer="+cfg.LrzUser,
		"-F", "mailadresse="+cfg.LrzMail,
		"-F", "telefon="+cfg.LrzPhone,
		"-F", "unidir=tum",
		"-F", "subdir="+cfg.LrzSubDir,
		"-F", "info=",
		cfg.LrzUploadUrl)
	_, err := cmd.CombinedOutput()
	if err != nil {
        return err
    }
	return nil
}

// PostFileUpload deprecated - example kindly provided by Attila O. Thanks buddy!
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
