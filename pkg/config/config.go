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
		ToFile        bool `yaml:"to_file"`
		ToConsole     bool `yaml:"to_console"`
		ToELK         bool `yaml:"to_elk"`
		UseZap        bool `yaml:"use_zap"`
		UseDefaultGin bool `yaml:"use_default_gin"`
		//IsJSON        bool `yaml:"is_Json"`
		//ToGraylog     bool   `yaml:"to_graylog"`
		//GraylogAddr   string `yaml:"graylog_addr"`
		//UseLogrus     bool   `yaml:"use_logrus"`
	} `yaml:"logger"`
}

const configPath = "app.yml"

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig(configPath, instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Fatalf("gelf.NewWriter: %s, %s", err, help)
		}
	})
	return instance
}
