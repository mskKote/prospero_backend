package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
)

// AppConfig - app.yml
type appConfig struct {
	Runtime           string `yaml:"runtime"`
	Service           string `yaml:"service" env-required:"true"`
	Port              string `yaml:"port" env-default:"5000"`
	CronSourcesRSS    string `yaml:"cron_sources_rss"`
	UseCronSourcesRSS bool   `yaml:"use_cron_sources_rss"`
	MigratePostgres   bool   `yaml:"migrate_postgres"`
	MigrateElastic    bool   `yaml:"migrate_elastic"`
	Metrics           bool   `yaml:"metrics"`
	UseTracingJaeger  bool   `yaml:"use_tracing_jaeger"`
	Logger            struct {
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

// Config - app.yml + .env
type Config struct {
	*appConfig
	Environment  string
	IsDebug      bool
	SecretKeyJWT string
	Tracing      struct {
		Host string
		Port string
	}
	Adminka struct {
		Username string
		Password string
	}
	Postgres struct {
		Username string
		Password string
		Host     string
		Port     string
		Database string
	}
	Elastic struct {
		Host string
		Port string
	}
}

const configPath = "app.yml"

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		// app.yml
		instanceApp := &appConfig{}
		if err := cleanenv.ReadConfig(configPath, instanceApp); err != nil {
			help, _ := cleanenv.GetDescription(instanceApp, nil)
			log.Fatalf("cleanenv: {%s}, {%s}", err, help)
		}
		instance.appConfig = instanceApp

		// .env
		if err := godotenv.Load(); err != nil {
			log.Fatal("Ошибка загрузки .env файла")
		}

		instance.SecretKeyJWT = getEnvKey("JWT_SECRET_KEY")

		instance.Tracing.Host = getEnvKey("JAEGER_HOST")
		instance.Tracing.Port = getEnvKey("JAEGER_PORT")

		instance.Postgres.Username = getEnvKey("POSTGRES_USERNAME")
		instance.Postgres.Password = getEnvKey("POSTGRES_PASSWORD")
		instance.Postgres.Host = getEnvKey("POSTGRES_HOST")
		instance.Postgres.Port = getEnvKey("POSTGRES_PORT")
		instance.Postgres.Database = getEnvKey("POSTGRES_DATABASE")

		instance.Elastic.Host = getEnvKey("ELASTIC_HOST")
		instance.Elastic.Port = getEnvKey("ELASTIC_PORT")

		instance.Adminka.Username = getEnvKey("ADMINKA_USERNAME")
		instance.Adminka.Password = getEnvKey("ADMINKA_PASSWORD")
		instance.Environment = getEnvKey("ENVIRONMENT")
		instance.IsDebug = instance.Environment != "production"
	})
	return instance
}

func getEnvKey(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Нет значения для " + key)
	return ""
}
