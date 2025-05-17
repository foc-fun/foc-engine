package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RpcConfig struct {
	Host string `yaml:"Host"`
}

type ApiConfig struct {
	Host         string   `yaml:"Host"`
	Port         int      `yaml:"Port"`
	AllowOrigins []string `yaml:"AllowOrigins"`
	AllowMethods []string `yaml:"AllowMethods"`
	AllowHeaders []string `yaml:"AllowHeaders"`
	Production   bool     `yaml:"Production"`
	Admin        bool     `yaml:"Admin"`
}

type IndexerConfig struct {
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`
}

type Config struct {
	Rpc     RpcConfig     `yaml:"Rpc"`
	Api     ApiConfig     `yaml:"Api"`
	Indexer IndexerConfig `yaml:"Indexer"`
}

var Conf *Config

func InitConfig() {
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		configPath = "configs/config.yaml"
		fmt.Println("CONFIG_PATH not set, using default config.yaml")
	}

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(yamlFile, &Conf)
	if err != nil {
		fmt.Println("Error parsing config file: ", err)
		os.Exit(1)
	}
}
