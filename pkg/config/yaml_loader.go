package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type ConfigYAML struct {
	Path string
}

func (yc *ConfigYAML) Load() (*Config, error) {
	c := &Config{}
	data, err := ioutil.ReadFile(yc.Path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
