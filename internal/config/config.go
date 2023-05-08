package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"sync"
)

type Config struct {
	Runtime string `yaml:"runtime"`
	IsDebug *bool  `yaml:"is_debug"`
	Listen  struct {
		Port string `yaml:"port" env-default:"5000"`
	} `yaml:"listen"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("Читаем app.yaml")
		instance = &Config{}
		if err := cleanenv.ReadConfig("app.yaml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatalln(err)
			panic(help)
		}
	})
	return instance
}
