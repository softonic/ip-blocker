package elasticsearch

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
)

var (
	stdlog, errlog *log.Logger
	r              map[string]interface{}
)

type ElasticSearch struct {
	client *elasticsearch.Client
}

func ElasticSearchInit(address string, username string, password string) ElasticSearch {

	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://host.docker.internal:9200", // Remember pass this to Env Variables
		},
		Username: username,
		Password: password, // Remember to pass this to secrets and not commit this to the repo
		//CACert:   cert,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS11,
				InsecureSkipVerify: true,
			},
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("\nError creating the client: ", err)
		os.Exit(1)
	}
	log.Println(elasticsearch.Version)

	return ElasticSearch{
		client: es,
	}

}

func ElasticSearchSearch(es ElasticSearch) map[string]int {

	//es := a.ElasticSearchClient

	client := es.client

	namespace := "istio-system"

	queryString := []byte(fmt.Sprintf(`{
  "query": {
    "bool": {
      "filter": [
          {
          "term": {
            "kubernetes.namespace": {
              "value": "%s"
            }
          }
        },
        {
          "match_phrase": {
            "response_code": "429"
          }
        },
        {
          "range": {
            "start_time": {
				"gt": "now-5m"
            }
          }
        }
      ]
    }
  }
}`, namespace))

	now := time.Now()

	botMap := make(map[string]int)

	read := bytes.NewReader(queryString)

	suffix := fmt.Sprintf("%d.%02d.%02d",
		now.Year(), now.Month(), now.Day())

	index := "istio-system-istio-system-" + suffix

	// Perform the search request.
	res, err := client.Search(
		//es.Search.WithContext(context.Background()),
		client.Search.WithIndex(index),
		client.Search.WithBody(read),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
		client.Search.WithSize(100),
	)
	if err != nil {
		log.Printf("error getting response: %s", err)
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		ips := hit.(map[string]interface{})["_source"].(map[string]interface{})["geoip"].(map[string]interface{})["ip"].(string)
		botMap[ips]++
	}

	return botMap

}
