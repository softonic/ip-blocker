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

	namespace := os.Getenv("NAMESPACE")

	intervalBlockTime, _ := strconv.Atoi(os.Getenv("INTERVAL_BLOCK_TIME"))

	ttlRules, _ := strconv.Atoi(os.Getenv("TTL_RULES"))

	threshold429PerMin, _ := strconv.Atoi(os.Getenv("THRESHOLD_429_PER_MIN"))

	flag.StringVar(&project, "project", "project", "kubernetes GCP project")
	flag.StringVar(&policy, "policy", "default", "The firewall rule that we will modify")

	flag.Parse()

	s := source.NewElasticSource(address, username, password, namespace, threshold429PerMin)
	a := actor.NewGCPArmorActor(project, policy)

	application := app.NewApp(s, a)

	application.Start(intervalBlockTime, ttlRules)

}
