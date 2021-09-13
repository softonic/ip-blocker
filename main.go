package main

import (
	_ "bytes"
	_ "io/ioutil"
	"log"
	"os"

	"github.com/softonic/ip-blocker/app"
	"github.com/softonic/ip-blocker/app/actor"
	"github.com/softonic/ip-blocker/app/source"
)

var (
	stdlog, errlog *log.Logger
)

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func main() {

	// GET THESE VARS FROM ENV

	address := os.Getenv("ELASTIC_ADDRESS")

	password := os.Getenv("ELASTIC_PASSWORD")

	username := os.Getenv("ELASTIC_USERNAME")

	project := "kubertonic"
	policy := "global-loadbalancer-rules"

	s := source.NewElasticSource(address, username, password)
	a := actor.NewGCPArmorActor(project, policy)

	application := app.NewApp(s, a)

	application.Start()

}
