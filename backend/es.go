/*
Package backend is the ElasticSearch backend manager and file retreival system.
*/
package backend

import (
	"fmt"
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"bytes"
	"html"

	elasticsearch "github.com/elastic/go-elasticsearch"
	esapi "github.com/elastic/go-elasticsearch/esapi"
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
	
	defer res.Body.Close()
	return fmt.Sprintf("%+v", res), nil
}

type dbRecord struct {
	Name string `json:"name"`
	Tags []string `json:"tags"`
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
	var buf bytes.Buffer
	esQuery := map[string]interface{}{
	  	"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": 	html.UnescapeString(query),
				"type": 	"best_fields",
		  		"fields": 	[]string{"name", "tags"},
			},
		  },
		  "size": 	25,
	}
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, fmt.Errorf("Error encoding query: %s", err)
	}

	// Perform the search request.
	res, err := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex("documents"),
		client.Search.WithBody(&buf),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	// Check if ElasticSearch returned an error
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("Error parsing the response body: %s", err)
		}
		// Print the response status and error information.
		return nil, fmt.Errorf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)
	}

	// Decode the response JSON
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