package main

import (
	"os"
	"sync"
	"time"

	"github.com/Montessquio/ChainDB/backend"
	"github.com/Montessquio/ChainDB/webui"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
    log.SetLevel(log.DebugLevel)
    
    log.Debug("Strike the Earth!")

    config := getCliArgs()
    log.WithField("data", config).Debug("CLI Arguments Processed")

    var wg sync.WaitGroup

    // Establish a connection to the ElasticSearch Server.
    // This will block until a connection is established.
    for {
        if resp, err := backend.InitES(config.esURL); err != nil {
            log.WithFields(log.Fields{
                "esURL": config.esURL,
                "Error": err,
            }).Info("Failed to connect to ElasticSearch.")
        } else {
            log.WithField("Response", resp).Info("Connected to ElasticSearch Database.")
            break
        }
        time.Sleep(2 * time.Second)
    }

    // Establish a connection to the backing Filesystem through Afero.
    backend.InitFS(config.storeDir)
    
    // Launch the Web UI
    wg.Add(1)
    go webui.Serve(config.port, config.siteRoot, config.sitePath, &wg)
        
    /*
    resp, err := backend.UpdateRecord(file.Name(), []string{"a", "b"})
    if err != nil {
        log.WithFields(log.Fields{
            "Error": err,
            "File": file.Name(),
            "Tags": []string{"a", "b"},
        }).Fatal("backend.UpdateRecord returned ERR")
    }
    log.WithFields(log.Fields{
        "File": file.Name(),
        "Tags": []string{"a", "b"},
        "Response": resp,
    }).Info("Updated Record in ES Database.")
    */
    
    wg.Wait()
}
