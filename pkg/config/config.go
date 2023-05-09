package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type Config struct {
	Runtime     string `yaml:"runtime"`
	Service     string `yaml:"service" env-required:"true"`
	GraylogAddr string `yaml:"graylog_addr" env-required:"true"`
	IsDebug     *bool  `yaml:"is_debug"`
	Listen      struct {
		Port string `yaml:"port" env-default:"5000"`
	} `yaml:"listen"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig("app.yaml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Fatalf("gelf.NewWriter: %s, %s", err, help)
		}
	})
	return instance
}
