package main

import (
	_ "bytes"
	"fmt"
	_ "io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	app "github.com/softonic/ip-blocker/app"
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


	application := app.NewApp()

	address := "https://host.docker.internal:9200"

	username := "elastic"
	password := ""

	es, err := application.ElasticSearchInit(address, username, password)
	if err != nil {
		errlog.Println("\nError search: ", err)
		os.Exit(1)
	}


	err = application.ElasticSearchSearch(es)
	if err != nil {
		errlog.Println("\nError search: ", err)
		os.Exit(1)
	}


	//cert, _ := ioutil.ReadFile("/tmp/cacert")


	service := NewService()
	//service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
