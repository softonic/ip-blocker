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
	bi             []app.Bot
	application    *app.App
)

// Service has embedded daemon
type Service struct {
	timeout time.Duration
}

/*
func NewService() *Service {
	listen := make(chan string)
	return &Service{
		listen: listen,
	}

}*/

// Manage by daemon commands or run the daemon
func (service *Service) Manage(listen chan []app.Bot, ips []string, lastPriority int32) (string, error) {

	// Do something, call your goroutines, etc

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	go doSearch(listen)

	for {
		select {
		case n := <-listen:
			fmt.Println(app.CompareBlockedIps(n, ips))
			fmt.Println(app.BlockedIPs(lastPriority))
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

func doSearch(listen chan []app.Bot) {
	for {
		time.Sleep(time.Millisecond * 300000)

		// we can call here ES connection and queries

		application.ElasticSearchSearch(listen)

	}
}

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)

}

func main() {

	application = app.NewApp()

	address := "https://host.docker.internal:9200"

	username := "elastic"
	password := "xxxx"

	err := application.ElasticSearchInit(address, username, password)
	if err != nil {
		errlog.Println("\nError search: ", err)
		os.Exit(1)
	}

	ips, lastPriority := app.ConnectGCP()

	listen := make(chan []app.Bot)

	service := Service{}

	//go application.ElasticSearchSearch(listen)
	status, err := service.Manage(listen, ips, lastPriority)
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
