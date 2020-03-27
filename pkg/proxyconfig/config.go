package proxyconfig

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	AutoSwitch bool
	PROXY      []ProxyConfig
}

type ProxyConfig struct {
	User        string
	Password    string
	AuthHost    string
	AuthPort    string
	LocalHost   string
	LocalPort   string
	Wifi        string
	Network     string
	Description string
}

func GetConfig(file string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
