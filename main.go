package main

import (
	_ "bytes"
	"flag"
	_ "io/ioutil"
	"log"
	"os"

	"github.com/softonic/ip-blocker/app"
	"github.com/softonic/ip-blocker/app/actor"
	"github.com/softonic/ip-blocker/app/source"
)

var (
	stdlog, errlog  *log.Logger
	project, policy string
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

	// GET THESE VARS FROM ARGS

	flag.StringVar(&project, "project", "project", "kubernetes GCP project")
	flag.StringVar(&policy, "policy", "default", "The firewall rule that we will modify")

	flag.Parse()

	s := source.NewElasticSource(address, username, password)
	a := actor.NewGCPArmorActor(project, policy)

	application := app.NewApp(s, a)

	application.Start()

}
