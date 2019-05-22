package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Addr string

	Auth struct {
		Key string
	}

	DB struct {
		Host     string
		Port     string
		Name     string
		User     string
		Password string
	}
}

func Load(filename string) (*Config, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err := yaml.Unmarshal(file, c); err != nil {
		return nil, err
	}

	return c, nil
}
