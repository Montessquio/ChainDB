package webui

import (
	"net/http"
	"sort"

	"github.com/Montessquio/ChainDB/backend"
	log "github.com/sirupsen/logrus"
)

// SiteData contains the data that is substituted into templates.
// This is instantiated per goroutine.
type siteData struct {
	// The site root, expressed in the form /x/y/z.
	// This is updated with the baseaddr. Does not contain a trailing slash.
	// Equal to basePath.
	Root string
	Query string
	NumResults int
	PaginationCookie string
	Results []backend.SearchResult
}


func webRootHandler(w http.ResponseWriter, r *http.Request) {
	templateData := siteData {
		Root: basePath,
	}

	// Fill templateData using the query string.
	// Only the first query will be used, the rest will be ignored.
	if val, ok := r.URL.Query()["q"]; ok {
		templateData.Query = val[0]
		// Query, then sort by Score
		results, err := backend.Search(templateData.Query)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err,
				"Query": templateData.Query,
			}).Error("Error searching ES Database with Query.")
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		templateData.Results = results
	} else {
		templateData.Query = ""
	}



	err := templates["main.html"].Execute(w, &templateData)
	if err != nil {
		log.WithFields(log.Fields{
			"Writer": w,
			"Request": r,
			"Error": err,
		}).Error("Error executing template.")
		http.Error(w, "", 500)
		return
	}
}