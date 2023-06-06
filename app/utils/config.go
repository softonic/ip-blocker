package utils

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Configurable interface {
	Parse([]byte) error
}

type MapConf struct {
	Data map[interface{}]interface{}
}

func (c *MapConf) Parse(data []byte) error {
	return yaml.Unmarshal(data, &c.Data)
}

func GetConf(filePath string, c Configurable) error {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("yamlFile.Get err   #%v ", err)
	}
	return c.Parse(yamlFile)
}

