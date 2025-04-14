package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type WgCtlConf struct {
	ServerCrt string `yaml:"server-crt"`
	ClientCrt string `yaml:"client-crt"`
	ClientKey string `yaml:"client-key"`
}

func LoadWgCliConf(cfgFile string) (*WgCtlConf, error) {
	f, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	cfg := new(WgCtlConf)
	if err := yaml.Unmarshal(f, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
