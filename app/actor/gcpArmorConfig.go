package actor

import (
	"os"

	"github.com/softonic/ip-blocker/app/utils"
	"gopkg.in/yaml.v2"
)

// ActorConfig is the configuration for the actor
type ActorConfig struct {
	Project    string
	Policy     string
	TTLRules   int
	ExcludeIPs string
}

// GCPArmorRulesConf is the configuration for the Armor rules
type GCPArmorRulesConf struct {
	preview bool   `yaml:"preview"`
	action  string `yaml:"action"`
}

func (c *GCPArmorRulesConf) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

var conf utils.MapConf

func (c *GCPArmorRulesConf) LoadConfig() error {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/config/gcp-armor-config.yaml" // valor por defecto
	}

	return utils.GetConf(configPath, &conf)

}
