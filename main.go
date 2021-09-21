package main

import (
	_ "bytes"
	"flag"
	_ "io/ioutil"
	"os"

	"k8s.io/klog"

	"github.com/softonic/ip-blocker/app"
	"github.com/softonic/ip-blocker/app/actor"
	"github.com/softonic/ip-blocker/app/source"
)

var (
	project, policy, namespace             string
	ttlRules, threshold, intervalBlockTime int
)

func init() {
	klog.InitFlags(nil)
}

func main() {

	// GET THESE VARS FROM ENV

	address := os.Getenv("ELASTIC_ADDRESS")

	password := os.Getenv("ELASTIC_PASSWORD")

	username := os.Getenv("ELASTIC_USERNAME")

	flag.StringVar(&project, "project", "project", "kubernetes GCP project")
	flag.StringVar(&policy, "policy", "default", "The firewall rule that we will modify")
	flag.StringVar(&namespace, "namespace", "istio-system", "elasticsearch k8s namespace to check 429s")
	flag.IntVar(&intervalBlockTime, "intervalBlockTime", 5, "check the 429s that we returned in the last N min")
	flag.IntVar(&ttlRules, "ttlRules", 60, "TTL in minutes of Firewall Rules. Once the ttl is exceeded, the rule is removed and the IPs are unblocked")
	flag.IntVar(&threshold, "threshold", 5, "we will check which IPs are being throttle , with a 429 code, per min, if exceed the threshold, there will be included in a blocked rule for at least ttlRules min")

	flag.Parse()

	s := source.NewElasticSource(address, username, password, namespace, threshold)
	a := actor.NewGCPArmorActor(project, policy, ttlRules)

	application := app.NewApp(s, a)

	application.Start(intervalBlockTime)

}
