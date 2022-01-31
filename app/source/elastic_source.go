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

	"gopkg.in/yaml.v2"

	"io/ioutil"

	"k8s.io/klog"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/softonic/ip-blocker/app"
)

var (
	r map[string]interface{}
)

type ElasticSource struct {
	client    *elasticsearch.Client
	threshold int
}

type ElasticConfig struct {
	Queries []struct {
		Name                 string `yaml:"name"`
		ElasticIndex         string `yaml:"elasticIndex"`
		ElasticFieldtoSearch string `yaml:"elasticFieldtoSearch"`
		QueryFile            string `yaml:"queryFile"`
	} `yaml:"queries"`
}

func NewElasticSource(address string, username string, password string, namespace string, threshold int, cacert string) *ElasticSource {
	return &ElasticSource{
		client:    elasticSearchInit(address, username, password, cacert),
		threshold: threshold,
	}
}

func elasticSearchInit(address string, username string, password string, cacert string) *elasticsearch.Client {

	var cert []byte

	if cacert != "" {
		cert, _ = ioutil.ReadFile(cacert)
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			address,
		},
		Username: username,
		Password: password,
		CACert:   cert,
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

// getElasticIndex returns the index name for the current day
func getElasticIndex(basename string) string {

	now := time.Now()

	suffix := fmt.Sprintf("%d.%02d.%02d",
		now.Year(), now.Month(), now.Day())

	index := basename + "-" + suffix

	return index

}

// orderAndTrimIPs returns a slice of IPs ordered by count
// limit the output to the top n
// limit the output to count > threshold
func orderAndTrimIPs(ipCounter map[string]int, maxCounter int) []app.IPCount {

	var ips []app.IPCount
	n := 10

	for ip, count := range ipCounter {
		if count > maxCounter {
			ips = append(ips, app.IPCount{
				IP:    ip,
				Count: int32(count),
			})
		}
	}

	sort.Slice(ips, func(i, j int) bool {
		return ips[i].Count > ips[j].Count
	})

	if len(ips) > n {
		ips = ips[:n]
	}

	return ips

}

func elasticSearchQuery(client *elasticsearch.Client, index string, query *bytes.Reader) (map[string]interface{}, error) {

	res, err := client.Search(
		client.Search.WithIndex(index),
		client.Search.WithBody(query),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
		client.Search.WithSize(10000),
	)

	if err != nil {
		return nil, err
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

	//r = res.Body

	return r, nil

}

func getQuery(file string) *bytes.Reader {

	jsonFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	queryString, _ := ioutil.ReadAll(jsonFile)

	read := bytes.NewReader(queryString)

	return read

}

func (s *ElasticSource) GetIPCount(interval int) []app.IPCount {

	//data := make(map[interface{}]string)

	var config ElasticConfig

	yamlFile, err := ioutil.ReadFile("/etc/config/elastic-search-config.yaml")
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	client := s.client

	threshold := s.threshold

	ipCounter := make(map[string]int)

	var listIPCandidates []app.IPCount

	//loop over all queries

	for _, query := range config.Queries {

		var listIPs []app.IPCount

		todayIndexName := getElasticIndex(query.ElasticIndex)
		read := getQuery(query.QueryFile)

		r, err := elasticSearchQuery(client, todayIndexName, read)
		if err != nil {
			klog.Errorf("\nError: %v", err)
			os.Exit(1)
		}

		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			ips := hit.(map[string]interface{})["_source"].(map[string]interface{})[query.ElasticFieldtoSearch].(map[string]interface{})["ip"].(string)
			ipCounter[ips]++
			klog.Infof("This is the ipcounter: %d", ipCounter)
		}

		maxCounter := interval * threshold

		klog.Infof("This is the counter: %d", maxCounter)

		listIPs = orderAndTrimIPs(ipCounter, maxCounter)

		klog.Infof("These are the listIPs after orderAndTrim: %v", listIPs)

		listIPCandidates = append(listIPCandidates, listIPs...)

	}

	return listIPCandidates

}
