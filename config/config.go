package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var Server *ServerCfg
var Client *ClientCfg
var Common *CommonCfg

type Config struct {
	ServerCfg `yaml:"server"`
	ClientCfg `yaml:"client"`
	CommonCfg `yaml:"common"`
}

type ServerCfg struct {
	Addr       string `yaml:"addr"`
	PrivateKey string `yaml:"privateKey"`
	CrtFile    string `yaml:"crtFile"`
	CAFile     string `yaml:"caFile"`
}

type ClientCfg struct {
	Addr       string `yaml:"addr"`
	PrivateKey string `yaml:"privateKey"`
	CrtFile    string `yaml:"crtFile"`
	Server     string `yaml:"server"`
}

type CommonCfg struct {
	Env     string `yaml:"env"`
	LogFile string `yaml:"logFile"`
}

func Init(file string) (err error) {
	viper.SetConfigFile(file)

	viper.AutomaticEnv() // read in environment variables that match

	viper.SetConfigType("yaml")

	// If a config file is found, read it in.
	if err = viper.ReadInConfig(); err != nil {
		return
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg, func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"
	})

	Server = &cfg.ServerCfg
	Client = &cfg.ClientCfg
	Common = &cfg.CommonCfg
	return
}
