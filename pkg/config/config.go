package config

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Accounts []Account
	Settings struct {
		CostCron      string `yaml:"costCron"`
		CoresInterval int64  `yaml:"coresInterval"`
	}
}

type Account struct {
	Name   string
	ID     string
	Secret string
}

func Parse(configPath string) Config {
	config := Config{}
	var err error

	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Fatal(err)
	}

	err = yaml.Unmarshal([]byte(raw), &config)
	if err != nil {
		logrus.Fatal(err)
	}

	return config
}
