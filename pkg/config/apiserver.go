package config

import (
	"os"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	cfgFile = "apiserver.yaml"
)

type ApiserverConf struct {
	BaseDir  string
	Port     int      `yaml:"port" default:"5443"`
	Key      string   `yaml:"key" default:"apiserver.key"`
	Crt      string   `yaml:"crt" default:"apiserver.crt"`
	DB       string   `yaml:"db" default:"apiserver.db"`
	AceessIP string   `yaml:"access-ip" validate:"required"`
	RRoutes  []string `yaml:"redirect-route"`
}

func LoadApiserverConf(dir string) (*ApiserverConf, error) {
	f, err := os.ReadFile(filepath.Join(dir, cfgFile))
	if err != nil {
		return nil, err
	}

	cfg := &ApiserverConf{
		BaseDir: dir,
	}
	if err := yaml.Unmarshal(f, cfg); err != nil {
		return nil, err
	}

	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	cfgValidator := validator.New()
	if err := cfgValidator.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
