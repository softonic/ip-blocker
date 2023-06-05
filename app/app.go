package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog"
)

type Source interface {
	GetIPCount(int) Result
}

type Actor interface {
	BlockIPs([]IPCount) error
	UnBlockIPs() error
}

type Result struct {
	IPCounts []IPCount
	Error    error
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

func getIPsToChannel(listen chan Result, source Source, interval int) {

	for {
		time.Sleep(time.Second * 60 * time.Duration(interval))

		listen <- source.GetIPCount(interval)

	}
}

func getBlockedIPsToChannel(exit chan bool, actor Actor) {
	for {
		time.Sleep(time.Second * 60 * 10)

		exit <- true

	}
}

func (a *App) Start(intervalBlockTime int) {

	klog.Infof("Starting daemon")

	listen := make(chan Result)
	exit := make(chan bool)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go getIPsToChannel(listen, a.source, intervalBlockTime)
	go getBlockedIPsToChannel(exit, a.actor)

	for {
		select {
		case result := <-listen:
			if result.Error != nil {
				klog.Errorf("\nError: %v", result.Error)
				continue
			}
			err := a.actor.BlockIPs(result.IPCounts)
			if err != nil {
				klog.Errorf("\nError: %v", err)
			}
		case <-exit:
			err := a.actor.UnBlockIPs()
			if err != nil {
				klog.Errorf("\nError: %v", err)
			}
		case killSignal := <-interrupt:
			klog.Infof("Got signal: %v", killSignal)
			klog.Infof("Stopping daemon")
		}
	}

}
