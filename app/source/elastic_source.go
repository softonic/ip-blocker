package source

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"k8s.io/klog"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/softonic/ip-blocker/app"
)

var (
	r map[string]interface{}
)

type ElasticSource struct {
	client    *elasticsearch.Client
	namespace string
	threshold int
}

func NewElasticSource(address string, username string, password string, namespace string, threshold int) *ElasticSource {
	return &ElasticSource{
		client:    elasticSearchInit(address, username, password),
		namespace: namespace,
		threshold: threshold,
	}
}

func elasticSearchInit(address string, username string, password string) *elasticsearch.Client {

	cfg := elasticsearch.Config{
		Addresses: []string{
			address,
		},
		Username: username,
		Password: password,
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
		klog.Errorf("\nError: %v", err)
		os.Exit(1)
	}

	klog.Infof("Connecting to ES in address: %s", address)

	klog.Infof("this is the ES version: %s", elasticsearch.Version)

	return es

}

func getElasticIndex(basename string) string {

	now := time.Now()

	suffix := fmt.Sprintf("%d.%02d.%02d",
		now.Year(), now.Month(), now.Day())

	index := basename + "-" + suffix

	return index

}

func (s *ElasticSource) GetIPCount(interval int) []app.IPCount {

	client := s.client

	namespace := s.namespace

	threshold := s.threshold

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
					"gt": "now-%dm"
				}
			  }
			}
		  ]
		}
	  }
	}`, namespace, interval))

	ipCounter := make(map[string]int)

	read := bytes.NewReader(queryString)

	todayIndexName := getElasticIndex("istio-system-istio-system")

	res, err := client.Search(
		client.Search.WithIndex(todayIndexName),
		client.Search.WithBody(read),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
		client.Search.WithSize(10000),
	)
	if err != nil {
		klog.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			klog.Fatalf("Error parsing the respnse body: %s", err)
		} else {
			klog.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)

		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		klog.Fatalf("Error parsing the response body: %s", err)
	}
	klog.Infof(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		ips := hit.(map[string]interface{})["_source"].(map[string]interface{})["geoip"].(map[string]interface{})["ip"].(string)
		ipCounter[ips]++
	}


	maxCounter := calculateCountBlockThreshold(threshold, interval)

	return orderAndTrimIPs(ipCounter, maxCounter)

}

func orderAndTrimIPs(ipCounter map[string]int, maxCounter int) []app.IPCount {

	bi := []app.IPCount{}
	output := []app.IPCount{}

	for ip, counter := range ipCounter {
		bot := app.IPCount{
			IP:    ip,
			Count: int32(counter),
		}
		if counter > maxCounter {
			bi = append(bi, bot)
		}

	}

	sort.Slice(bi, func(i, j int) bool {
		return bi[i].Count > bi[j].Count
	})

	if len(bi) > 10 {
		for i := 0; i < 10; i++ {
			output = append(output, bi[i])
		}
	} else {
		for i := 0; i < len(bi); i++ {
			output = append(output, bi[i])
		}
	}

	return output

}

func calculateCountBlockThreshold(threshold429PerMin int, interval int) int {

	return interval * threshold429PerMin

}
