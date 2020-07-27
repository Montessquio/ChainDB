/*

Parse and validate command-line flags.

*/
package main

import (
	"flag"
	"net/url"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type cliArgs struct {
    port        int
    storeDir    string
	siteRoot    string
	sitePath	string
	esURL		string
}

// Gets and validates command-line flags.
func getCliArgs() cliArgs {
	// Define flags
	port := flag.Int("p", 80, "Port to serve on")
	storeDir := flag.String("d", "", "Data directory to serve files from.")
	siteRoot := flag.String("r", "", "The domain that this utility is served under. Valid values include 192.168.2.1:80, app.example.net, and example.net")
	sitePath := flag.String("s", "", "The subdirectory of the domain this utility will be served under. If this is set to \"db\" then the page found normally at example.com/a.html will instead be found at example.com/db/a.html")
	esURL := flag.String("e", "", "The URL of the ElasticSearch server.")

	flag.Parse()

	// Assign and check flags.
	args := cliArgs{}

	args.port = *port
	if args.port > 65535 || args.port < 1 {
		log.WithFields(log.Fields{
			"port": args.port,
		}).Fatal("Port out of range")
	}

	if len(*storeDir) == 0 {
		log.Fatal("Directory store location required.")
	}
	args.storeDir = path.Clean(*storeDir)
	_, err := os.Stat(args.storeDir)
    if os.IsNotExist(err) {
        log.WithFields(log.Fields{
			"dir": args.storeDir,
		}).Fatal("Folder does not exist.")
    }
	
	if len(*siteRoot) == 0 {
		log.Warnf("Site root not set, defaulting to localhost:%s", strconv.Itoa(args.port))
		args.siteRoot = "localhost:" + strconv.Itoa(args.port)
	} else {
		args.siteRoot = *siteRoot
		_, err = url.ParseRequestURI(args.siteRoot)
		if err != nil {
			log.WithFields(log.Fields{
				"url": args.siteRoot,
			}).Fatal("The Site Root must be a valid URL.")
		}
	}

	args.sitePath = *sitePath

	if len(*esURL) == 0 {
		log.Fatal("ElasticSearch URL required.")
	}
	args.esURL = *esURL
	_, err = url.ParseRequestURI(args.esURL)
	if err != nil {
		log.WithFields(log.Fields{
			"url": args.esURL,
		}).Fatal("The ElasticSearch URL must be a valid URL.")
	}

	return args
}