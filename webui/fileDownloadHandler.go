package webui

import (
	"net/http"
	"net/url"

	"github.com/Montessquio/ChainDB/backend"

	log "github.com/sirupsen/logrus"
)

// Catches any request to ^/file/(.*)$ and returns the Backend result for that file.
func fileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	fileURL, err := url.QueryUnescape(r.URL.Path[len(setpath("/file/"))-1:])
	if err != nil {
		log.WithFields(log.Fields{
			"Writer": w,
			"Request": r,
			"Error": err,
		}).Error("Error unescaping file from from frontend.")
		http.Error(w, "", 500)
	}

	file, err := backend.OpenFile(fileURL)
	if err != nil {
		log.WithFields(log.Fields{
			"Writer": w,
			"Request": r,
			"Error": err,
		}).Error("Error opening file from backend.")
		http.Error(w, "", 500)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.WithFields(log.Fields{
			"Writer": w,
			"Request": r,
			"Error": err,
		}).Error("Error on file stat from backend.")
		http.Error(w, "", 500)
		return
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}