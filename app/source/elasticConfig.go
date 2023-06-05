package source

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/softonic/ip-blocker/app/utils"
)

// This struct is used to allocate the configuration of the ElasticSearch connection
type SourceConfig struct {
	Address   string
	Username  string
	Password  string
	Namespace string
	Threshold int
	CACert    string
}

// This struct is used to parse the yaml file with the queries to execute
type ElasticQueryConfig struct {
	Queries []struct {
		Name                 string `yaml:"name"`
		ElasticIndex         string `yaml:"elasticIndex"`
		ElasticFieldtoSearch string `yaml:"elasticFieldtoSearch"`
		QueryFile            string `yaml:"queryFile"`
	} `yaml:"queries"`
}

func (c *ElasticQueryConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func (c *ElasticQueryConfig) LoadConfig() error {
	configPath := os.Getenv("ES_CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/config/elastic-search-config.yaml" // valor por defecto
	}

	return utils.GetConf(configPath, c)

}
