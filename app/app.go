package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	stdlog *log.Logger
)

type Source interface {
	GetIPCount() []IPCount
}

type Actor interface {
	BlockIPs([]IPCount) error
}

type IPCount struct {
	IP    string
	Count int32
}

// App : Basic struct
type App struct {
	source Source
	actor  Actor
}

func NewApp(s Source, a Actor) *App {

	return &App{
		source: s,
		actor:  a,
	}
}

func getIPsToChannel(listen chan []IPCount, source Source) {
	for {
		time.Sleep(time.Millisecond * 300000)

		listen <- source.GetIPCount()

	}
}

func (a *App) Start() {

	listen := make(chan []IPCount)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go getIPsToChannel(listen, a.source)

	for {
		select {
		case sourceIPs := <-listen:
			err := a.actor.BlockIPs(sourceIPs)
			fmt.Println(err)
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			stdlog.Println("Stoping daemon")
		}
	}

}
