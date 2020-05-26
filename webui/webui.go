/*
Package webui presents a pretty web UI to search, upload and download to/from the database.
*/
package webui

import (
	"sync"
	"net/http"
	"net/url"
	"path"

	"strconv"

	log "github.com/sirupsen/logrus"
    //"github.com/gorilla/securecookie"
)

var basePath string
var baseAddr string

// Serve the web frontend on the given port.
func Serve(port int, bAddr string, bPath string, wg *sync.WaitGroup) {
	defer wg.Done()

	/* Cache essential variables and templates */
	baseAddr = bAddr
	basePath = bPath
	
	initTemplates()

	log.WithFields(log.Fields{
		"Port": port,
		"BaseAddr": baseAddr,
		"BasePath": basePath,
	}).Info("Launching web frontend")

	/* Set request Handlers */
	http.HandleFunc(setpath("/"), webRootHandler)
	log.Debugf("Set handler function for location %s", setpath("/"))

	http.HandleFunc(setpath("/file/"), fileDownloadHandler)
	log.Debugf("Set handler function for location %s", setpath("/file/"))

	// Serve static files from the web directory.
	http.Handle(	
		setpath("/static/"), http.StripPrefix(setpath("/static/"), 
		http.FileServer(http.Dir("www/static/"))),
	)
	log.Debugf("Set handler function for location %s", setpath("/static/"))

	/* Start the Server */
	log.WithFields(log.Fields{
		"Base Addr": baseAddr,
		"Base Path": basePath,
		"Port": port,
	}).Info("Serving...")
	log.Error(http.ListenAndServe(":" + strconv.Itoa(port), nil))
}

// Function to easily and safely prepend the base path to a given URL path.
func setpath(p string) string {
	u, err := url.Parse(basePath)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"BasePath": basePath,
			"path": p,
		}).Error("Error parsing BasePath")
		return p
	}
	// path.Join removes trailing slashes, so we must append one here.
	u.Path = path.Join(u.Path, p) + "/"
	return u.String()
}