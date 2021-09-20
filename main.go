package main

import (
	_ "bytes"
	"flag"
	_ "io/ioutil"
	"os"
	"strconv"

	"k8s.io/klog"

	"github.com/softonic/ip-blocker/app"
	"github.com/softonic/ip-blocker/app/actor"
	"github.com/softonic/ip-blocker/app/source"
)

var (
	project, policy string
)

func init() {
	klog.InitFlags(nil)
}

func main() {

	// GET THESE VARS FROM ENV

	address := os.Getenv("ELASTIC_ADDRESS")

	password := os.Getenv("ELASTIC_PASSWORD")

	username := os.Getenv("ELASTIC_USERNAME")

	intervalBlockTime, _ := strconv.Atoi(os.Getenv("INTERVAL_BLOCK_TIME"))

	intervalUnBlockTime, _ := strconv.Atoi(os.Getenv("INTERVAL_UNBLOCK_TIME"))

	flag.StringVar(&project, "project", "project", "kubernetes GCP project")
	flag.StringVar(&policy, "policy", "default", "The firewall rule that we will modify")

	flag.Parse()

	s := source.NewElasticSource(address, username, password)
	a := actor.NewGCPArmorActor(project, policy)

	application := app.NewApp(s, a)

	application.Start(intervalBlockTime, intervalUnBlockTime)

}
