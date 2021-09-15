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
	UnBlockIPs() error
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
		time.Sleep(time.Millisecond * 100000)

		listen <- source.GetIPCount()

	}
}

func getBlockedIPsToChannel(exit chan bool, actor Actor) {
	for {
		time.Sleep(time.Millisecond * 300000)

		//blocked <- actor.GetBlockedIPsFromActorThatCanBeUnblocked()

		exit <- true

	}
}

func (a *App) Start() {

	listen := make(chan []IPCount)
	//blocked := make(chan []string)
	exit := make(chan bool)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go getIPsToChannel(listen, a.source)
	go getBlockedIPsToChannel(exit, a.actor)

	for {
		select {
		case sourceIPs := <-listen:
			err := a.actor.BlockIPs(sourceIPs)
			fmt.Println(err)
		case <-exit:
			//fmt.Println("these are the blockedIPs:", blockedIPs)
			err := a.actor.UnBlockIPs()
			fmt.Println(err)
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			stdlog.Println("Stoping daemon")
		}
	}

}
