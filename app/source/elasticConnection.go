package source

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"k8s.io/klog"
)

type ElasticConnection struct {
	client    *elasticsearch.Client
	threshold int
}

func NewElasticConnection(address string, username string, password string, threshold int, cacert string) (*ElasticConnection, error) {

	client, err := elasticSearchInit(address, username, password, cacert)
	if err != nil {
		klog.Error("\nError: ", err)
		return nil, err
	}

	return &ElasticConnection{
		client:    client,
		threshold: threshold,
	}, nil

}

func elasticSearchInit(address string, username string, password string, cacert string) (*elasticsearch.Client, error) {

	var cert []byte

	if cacert != "" {
		cert, _ = ioutil.ReadFile(cacert)
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			address,
		},
		Username: username,
		Password: password,
		CACert:   cert,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS11,
				InsecureSkipVerify: true,
			},
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		klog.Errorf("\nError: %v", err)
		return nil, err
	}

	klog.Infof("Connecting to ES in address: %s", address)

	klog.Infof("this is the ES version: %s", elasticsearch.Version)

	return es, err

}
