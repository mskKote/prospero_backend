package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type Config struct {
	Runtime string `yaml:"runtime"`
	Service string `yaml:"service" env-required:"true"`
	Port    string `yaml:"port" env-default:"5000"`
	IsDebug bool   `yaml:"is_debug"`
	Logger  struct {
		ToGraylog     bool   `yaml:"to_graylog"`
		GraylogAddr   string `yaml:"graylog_addr"`
		ToFile        bool   `yaml:"to_file"`
		ToConsole     bool   `yaml:"to_console"`
		IsJSON        bool   `yaml:"is_Json"`
		UseDefaultGin bool   `yaml:"use_default_gin"`
	} `yaml:"logger"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig("app.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Fatalf("gelf.NewWriter: %s, %s", err, help)
		}
	})
	return instance
}
