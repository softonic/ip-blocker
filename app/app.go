package app

import (
	"fmt"
	"log"
	"sync"

	"github.com/softonic/ip-blocker/app/actor"
	"github.com/softonic/ip-blocker/app/source"
)

var (
	stdlog, errlog *log.Logger
	r              map[string]interface{}
	wg             sync.WaitGroup
)

// App : Basic struct
type App struct {
	Source source.Source
	Actor  actor.Actor
}

type Bot struct {
	ip      string
	count   int
	blocked bool
}

func NewBot() *Bot {
	return &Bot{}
}

func NewApp() *App {

	source := source.Source{}
	actor := actor.Actor{}

	return &App{
		Source: source,
		Actor:  actor,
	}
}

func (a *App) InitSource(address string, username string, password string) {
	a.Source.InitConnectiontoSource(address, username, password)
}

func (a *App) InitActor() {
	a.Actor.InitConnectiontoActor()
}

func (a *App) SearchSource(listen chan []Bot) error {

	bi := []Bot{}

	botMap := a.Source.SearchSource()

	for i, k := range botMap {
		bot := Bot{
			ip:    i,
			count: k,
		}
		bi = append(bi, bot)
	}

	fmt.Println("These are the bots from ES:", bi)

	listen <- bi

	return nil

}

func (a *App) GetInfoActor() ([]string, int32) {

	ips, lastPriority := a.Actor.GetIPsfromRulesActor()

	return ips, lastPriority

}

func (a *App) BlockIps(prio int32, candidateIPsBlocked []string) (error, string) {

	err, status := a.Actor.BlockIPsFromActor(prio, candidateIPsBlocked)

	return err, status

}

func CompareBlockedIps(bots []Bot, ips []string) []string {

	// compare the array of IPs of ES with the IPs of GCP armor

	fmt.Println("Ips received from ES:", bots)
	fmt.Println("IPs armor received in the func:", ips)

	var count int
	var ipWithMaskES string
	candidateIPsBlocked := []string{}

	for _, elasticIps := range bots {
		for _, armorIps := range ips {
			ipWithMaskES = elasticIps.ip + "/32"
			if ipWithMaskES == armorIps {
				fmt.Println("this IPs are already in the armor")
				count++
			}
		}
		fmt.Println("this IP is not in armor:", ipWithMaskES)
		candidateIPsBlocked = append(candidateIPsBlocked, ipWithMaskES)
	}

	return candidateIPsBlocked

}
