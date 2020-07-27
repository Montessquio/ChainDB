/*
Package backend is the ElasticSearch backend manager and file retreival system.
*/
package backend

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch"
	esapi "github.com/elastic/go-elasticsearch/esapi"
	log "github.com/sirupsen/logrus"
)

var client *elasticsearch.Client;

// InitES pings the ES Server.
// Returns an error if the connection did not succeed.
func InitES(URL string) (string, error) {
	cfg := elasticsearch.Config {
		Addresses: []string{
			URL,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
	  return "", fmt.Errorf("Error creating the client: %s", err.Error())
	}
	client = es
	
	res, err := client.Info()
	if err != nil {
	   return "", fmt.Errorf("Error getting response: %s", err.Error())
	}

	err = initIndex()
	if err != nil {
		return "", fmt.Errorf("Error Initializing Index: %s", err.Error())
	}
	
	defer res.Body.Close()
	return fmt.Sprintf("%+v", res), nil
}

type dbRecord struct {
	Name string `json:"name"`
	Tags []string `json:"tags"`
}

// Creates the index if it doesn't exist.
func initIndex() error {
	// Create index if it doesn't exist.
	req := 
	esapi.IndicesCreateRequest{
		Index: "documents",
		Body: strings.NewReader(`{
			"settings": {
				"max_ngram_diff": 8,
				"analysis": {
					"analyzer": {
					  "pdf_analyzer": {
						"type": "custom",
						"tokenizer": "pdf_ngram_tokenizer",
						"filter": [
						  "pdf_ngram_filter"
						]
					  }
					},
					"filter": {
					  "pdf_ngram_filter": {
						"type": "ngram",
						"min_gram": 2,
						"max_gram": 8
					  }
					},
					"tokenizer": {
					  "pdf_ngram_tokenizer": {
						"type": "ngram",
						"min_gram": 2,
						"max_gram": 8
					  }
					}
				}
			}
		}`),
	}

	// Actually perform the request.
	res, err := req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("Error performing ESAPI index request: %s", err.Error())
	}
	defer res.Body.Close()

	// This means it either already exists, or something else happened.
	if res.IsError() {
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			return fmt.Errorf("Error parsing the response body: %s", err.Error())
		}

		// This horrible thing extracts some VERY deeply nested JSON, to find which exception was raised by the
		// ES query that attempted to create the index.
		// If the exeption was that the index already exists, we can safely continue.
		// Otherwise, we send the entire response body back as an error.
		if r["error"].(map[string]interface{})["root_cause"].([]interface{})[0].(map[string]interface{})["type"].(string) == "resource_already_exists_exception" {
			log.Info("Index `documents` already exists. Continuing...")
			return nil
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		return fmt.Errorf("Got Response Error from ES Database: %s, %s", res.Status(), buf.String())
	}

	log.Warn("Created new Index")
	return nil
}

// UpdateRecord modifies the tags associated with a particular filename in the ES database.
//
// Internally, the key used is the Base64 representation of the filename.
func UpdateRecord(fileName string, tags []string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("Client is nil. Client must be initialized with backend.InitES(URL string)")
	}

	// Create the base64 key.
	key := base64.StdEncoding.EncodeToString([]byte(fileName))

	// Establish the complete ES record and marshal it to JSON
	esBody, err := json.Marshal(dbRecord{ Name: fileName, Tags: tags,})
	if err != nil {
		return "", fmt.Errorf("Error marshalling DB Record to JSON: %s", err.Error())
	}

	req := esapi.IndexRequest{
		Index: 		"documents",
		DocumentID: key,
		Body: 		strings.NewReader(string(esBody)),
		Refresh: 	"true",
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return "", fmt.Errorf("Error performing ESAPI index request: %s", err.Error())
	}
	defer res.Body.Close()

	if res.IsError() {
    	return "", fmt.Errorf("Got Response Error from ES Database: %s", res.Status())
	}

	// Deserialize the response into a map.
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("Error parsing the response body: %s", err.Error())
	}

	return fmt.Sprintf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64))), nil

}

// DeleteRecord removes a record from the ES database but not its corresponding file in the filesystem.
func DeleteRecord(fileName string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("Client is nil. Client must be initialized with backend.InitES(URL string)")
	}
	
	key := base64.StdEncoding.EncodeToString([]byte(fileName))

	req := esapi.DeleteRequest{
		Index: "documents",
		DocumentID: key,
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return "", fmt.Errorf("Error performing ESAPI delete request: %s", err.Error())
	}
	defer res.Body.Close()

	if res.IsError() {
    	return "", fmt.Errorf("Got Response Error from ES Database: %s", res.Status())
	}

	// Deserialize the response into a map.
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("Error parsing the response body: %s", err.Error())
	}

	return fmt.Sprintf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64))), nil
}

// SearchResult comprises a single result from the ElasticSearch index.
type SearchResult struct {
	FileName string
	FileTags []string
	Score float64
}

// Search for a given query in the ES index and return a slice of results.
func Search(query string) ([]SearchResult, error) {
	if client == nil {
		return nil, fmt.Errorf("Client is nil. Client must be initialized with backend.InitES(URL string)")
	}

	// Construct the request body.
	// This runs the query through a search in the name and tags fields.
	buf, err := createQueryJSON(query)
	if err != nil {
		return nil, err
	}

	// Perform the search request.
	res, err := performQuery(buf)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode and return response JSON (or any errors)
	return parseQueryResponse(res)
}

// Creates a JSON string, ready to pass to an ES Query.
// This is currently non-functional. The query construction
// process must be carefully refactored to produce better
// search results.
func createQueryJSON(query ...string) (bytes.Buffer, error) {
	// Construct the request body.
	// This runs the query through a search in the name and tags fields.
	var buf bytes.Buffer
	esQuery := map[string]interface{}{
	  	"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": 	html.UnescapeString(query[0]),
				"type": 	"best_fields",
		  		"fields": 	[]string{"name", "tags"},
			},
		  },
		  "size": 	25,
	}
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return bytes.Buffer{}, fmt.Errorf("Error encoding query: %s", err)
	}
	return buf, nil
}

// Performs a search request for a document based on the
// request in buf, and checks for any errors down the line.
func performQuery(buf bytes.Buffer) (*esapi.Response, error) {
	// Perform the search request.
	res, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex("documents"),
		client.Search.WithBody(&buf),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("Error performing ES Search: %s", err)
	}

	// Check if ElasticSearch returned an error
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("Error parsing the error response body: %s", err)
		}
		// Print the response status and error information.
		return nil, fmt.Errorf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)
	}

	return res, nil
}

// Converts the ElasticSearch raw JSON response into a SearchResult slice,
// ready for use in the HTML Template Engine.
func parseQueryResponse(res *esapi.Response) ([]SearchResult, error) {
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("Error parsing the response body: %s", err)
	}

	// Save the name and tags for each hit.
	var results []SearchResult
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		fName := hit.(map[string]interface{})["_source"].(map[string]interface{})["name"].(string)
		var fTags []string
		for _, val := range hit.(map[string]interface{})["_source"].(map[string]interface{})["tags"].([]interface{}) {
			fTags = append(fTags, val.(string))
		}
		fScore := hit.(map[string]interface{})["_score"].(float64)

		results = append(results, SearchResult{
			FileName: fName,
			FileTags: fTags,
			Score: fScore,
		})
	}
	return results, nil
}