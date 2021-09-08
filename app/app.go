package app

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	_ "fmt"
	"log"
	"os"
	"sync"
	"time"

	//"sync"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
)

var (
	stdlog, errlog *log.Logger
	r              map[string]interface{}
	wg             sync.WaitGroup
)

// App : Basic struct
type App struct {
	GCPConnection
	ElasticSearchClient *elasticsearch.Client
}

type GCPConnection struct{}

type Bot struct {
	ip                string
	count             int
	lastReadTimestamp int64
	blocked           bool
}

func NewBot() *Bot {
	return &Bot{}
}

func NewApp() *App {

	return &App{
		ElasticSearchClient: &elasticsearch.Client{},
		GCPConnection:       GCPConnection{},
	}
}

func (a *App) ElasticSearchInit(address string, username string, password string) error {

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
		errlog.Println("\nError creating the client: ", err)
		os.Exit(1)
	}
	log.Println(elasticsearch.Version)

	a.ElasticSearchClient = es

	return err

}

func (a *App) ElasticSearchSearch(listen chan []Bot) error {

	bi := []Bot{}

	es := a.ElasticSearchClient

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
	secs := now.Unix()

	botMap := make(map[string]int)

	read := bytes.NewReader(queryString)

	suffix := fmt.Sprintf("%d.%02d.%02d",
		now.Year(), now.Month(), now.Day())

	index := "istio-system-istio-system-" + suffix

	// Perform the search request.
	res, err := es.Search(
		//es.Search.WithContext(context.Background()),
		es.Search.WithIndex(index),
		es.Search.WithBody(read),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
		es.Search.WithSize(100),
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
			// Print the response status and error information.
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

	for i, k := range botMap {
		bot := Bot{
			ip:                i,
			count:             k,
			lastReadTimestamp: secs,
		}
		bi = append(bi, bot)
	}

	listen <- bi

	return nil

}

func CompareBlockedIps(bots []Bot, ips []string) string {

	// compare the array of IPs of ES with the IPs of GCP armor

	fmt.Println("Ips received from ES:", bots)
	fmt.Println("IPs armor received in the func:", ips)

	var count int
	var ipWithMask string

	for _, elasticIps := range bots {
		for _, armorIps := range ips {
			ipWithMask = elasticIps.ip + "/32"
			if ipWithMask == armorIps {
				fmt.Println("this IPs are already in the armor")
				count++
			}
		}
		fmt.Println("this IP is not in armor:", ipWithMask)
	}

	if count > 0 {
		return "there is one IP that it is already in GCP"
	}

	return "this IP is not in armor"

}
