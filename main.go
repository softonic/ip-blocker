package main

import (
	_ "bytes"
	"errors"
	"flag"
	_ "io/ioutil"
	"os"

	"k8s.io/klog"

	"github.com/softonic/ip-blocker/app"
	"github.com/softonic/ip-blocker/app/actor"
	"github.com/softonic/ip-blocker/app/source"
)

var (
	project, policy, cacert, excludeIPs    string
	ttlRules, threshold, intervalBlockTime int
)

func init() {
	klog.InitFlags(nil)
}

func main() {

	// GET THESE VARS FROM ENV

	flag.StringVar(&project, "project", "project", "kubernetes GCP project")
	flag.StringVar(&policy, "policy", "default", "The firewall rule that we will modify")
	flag.IntVar(&intervalBlockTime, "intervalBlockTime", 1, "check the 429s that we returned in the last N min")
	flag.IntVar(&ttlRules, "ttlRules", 60, "TTL in minutes of Firewall Rules. Once the ttl is exceeded, the rule is removed and the IPs are unblocked")
	flag.IntVar(&threshold, "threshold", 5, "we will check which IPs are being throttle , with a 429 code, per min, if exceed the threshold, there will be included in a blocked rule for at least ttlRules min")
	flag.StringVar(&cacert, "cacert", "", "If you are connecting to a ES that needs TLS, this is the ca certificate")
	flag.StringVar(&excludeIPs, "excludeIPs", "", "comma separeted IPs that will be excluded from blocker, e.g., 1.1.1.1, 2.2.2.2")

	flag.Parse()

	sourceConfig, err := getElasticSourceConfigFromEnv()
	if err != nil {
		klog.Fatalf("Failed to get config from env: %v", err)
		os.Exit(1)
	}

	actorConfig, err := getActorConfigFromEnv()
	if err != nil {
		klog.Fatalf("Failed to get config from env: %v", err)
		os.Exit(1)
	}

	// create the source. If cannot be created, exit
	s, err := source.NewElasticSource(sourceConfig)
	if err != nil {
		klog.Fatalf("Failed to create ElasticSource: %v", err)
		os.Exit(1)
	}
	// create the actor. If cannot be created, exit
	a, err := actor.NewGCPArmorActor(actorConfig)
	if err != nil {
		klog.Fatalf("Failed to create GCPArmorActor: %v", err)
		os.Exit(1)
	}

	application := app.NewApp(s, a)

	application.Start(intervalBlockTime)

}

func getElasticSourceConfigFromEnv() (*source.SourceConfig, error) {
	// This assumes you have a structure named Config in the app package that holds
	// all the necessary configuration information.
	config := &source.SourceConfig{}

	config.Address = os.Getenv("ELASTIC_ADDRESS")
	if config.Address == "" {
		return nil, errors.New("ELASTIC_ADDRESS environment variable not set")
	}
	config.Password = os.Getenv("ELASTIC_PASSWORD")
	if config.Address == "" {
		return nil, errors.New("ELASTIC_ADDRESS environment variable not set")
	}
	config.Username = os.Getenv("ELASTIC_USERNAME")
	if config.Address == "" {
		return nil, errors.New("ELASTIC_ADDRESS environment variable not set")
	}
	config.Threshold = threshold
	config.CACert = cacert

	// Remember to check if the necessary env variables were set,
	// returning an error if they weren't

	return config, nil
}

func getActorConfigFromEnv() (*actor.ActorConfig, error) {
	// This assumes you have a structure named Config in the app package that holds
	// all the necessary configuration information.
	config := &actor.ActorConfig{}

	config.Project = project
	config.Policy = policy
	config.TTLRules = ttlRules
	config.ExcludeIPs = excludeIPs

	// Remember to check if the necessary env variables were set,
	// returning an error if they weren't

	return config, nil
}
