package main

import (
	_ "bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	_ "io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

//    dependencies that are NOT required by the service, but might be used

var (
	stdlog, errlog *log.Logger
	r              map[string]interface{}
	wg             sync.WaitGroup
)

// Service has embedded daemon
type Service struct {
	listen chan string
}

type Bot struct {
	ip string
	count int
	lastReadTimestamp int64
	blocked bool
}

type ArmorRules struct {
	CreationTimestamp string `json:"creationTimestamp"`
	Description       string `json:"description"`
	Fingerprint       string `json:"fingerprint"`
	ID                string `json:"id"`
	Kind              string `json:"kind"`
	Name              string `json:"name"`
	Rules             []struct {
		Action      string `json:"action"`
		Description string `json:"description"`
		Kind        string `json:"kind"`
		Match       struct {
			Config struct {
				SrcIPRanges []string `json:"srcIpRanges"`
			} `json:"config"`
			VersionedExpr string `json:"versionedExpr"`
		} `json:"match"`
		Preview  bool `json:"preview"`
		Priority int  `json:"priority"`
	} `json:"rules"`
	SelfLink string `json:"selfLink"`
}


func NewService() *Service {
	listen := make(chan string)
	return &Service{
		listen: listen,
	}

}

func NewBot() *Bot {
	return &Bot{}
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	// Do something, call your goroutines, etc

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	go printInfinite(service.listen)

	// set up channel on which to send accepted connections

	// loop work cycle with accept connections or interrupt
	// by system signal
	for {
		select {
		case n := <-service.listen:
			fmt.Println(n)
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			stdlog.Println("Stoping daemon")
			if killSignal == os.Interrupt {
				return "Daemon was interruped by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}

	// never happen, but need to complete code
	return "", nil
}

func printInfinite(ch chan string) {
	for {
		time.Sleep(time.Millisecond * 400000)
		ch <- fmt.Sprintf("hello")
		// we can call here ES connection and queries
	}
}

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)

}

func main() {

	//cert, _ := ioutil.ReadFile("/tmp/cacert")

	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://host.docker.internal:9200",		// Remember pass this to Env Variables
		},
		Username: "",
		Password: "",						// Remember to pass this to secrets and not commit this to the public repo
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
	/*
		res, err := es.Info()
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}

		defer res.Body.Close()
		log.Println(res)

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}
		// Print client and server version numbers.
		log.Printf("Client: %s", elasticsearch.Version)
		log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
		log.Println(strings.Repeat("~", 37))*/



	queryString := `{
  "query": {
    "bool": {
      "filter": [
          {
          "term": {
            "kubernetes.namespace": {
              "value": "istio-system"
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
              "gte": "2021-08-12T09:28:10.075Z",
              "lte": "2021-08-12T10:28:10.075Z",
              "format": "strict_date_optional_time"
            }
          }
        }
      ]
    }
  }
}`


	now := time.Now()
	secs := now.Unix()

	var bots []Bot
	botMap := make(map[string]int)
	var b strings.Builder
	//var res *esapi.Response
	b.WriteString(queryString)
	read := strings.NewReader(b.String())

	fmt.Printf("%d.%02d.%02d\n",
		now.Year(), now.Month(), now.Day())

	// Perform the search request.
	res, err := es.Search(
		//es.Search.WithContext(context.Background()),
		es.Search.WithIndex("istio-system-istio-system-2021.08.12"),
		es.Search.WithBody(read),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
		es.Search.WithSize(200),
	)
	if err != nil {
		log.Printf("error getting response: %s", err)
		log.Fatalf("Error getting response: %s", err)
	}

	//defer res.Body.Close()

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
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the IP from _source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		//log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		//log.Printf("IP: %s", hit.(map[string]interface{})["_source"].(map[string]interface{})["geoip"].(map[string]interface{})["ip"])
		ips := hit.(map[string]interface{})["_source"].(map[string]interface{})["geoip"].(map[string]interface{})["ip"].(string)
		//log.Printf("IP: %s", ips)

		botMap[ips] ++
	}


	for i,k := range botMap {
		bot := Bot{
			ip: i,
			count: k,
			lastReadTimestamp: secs,
		}
		bots = append(bots, bot)
	}

	//log.Printf("IP: %s, %d", bots[1].ip, bots[1].count)
	//log.Println(botMap)
	log.Printf("Bots: %+v\n", bots)
	log.Println(strings.Repeat("=", 37))


	res.Body.Close()
	log.Println("closing conn")

	service := NewService()
	//service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
