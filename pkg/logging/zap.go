package logging

import (
	"github.com/mskKote/prospero_backend/pkg/config"
	"go.uber.org/zap"
	"log"
	"os"
	"sync"
)

// ----------------------------- Setup
var cfg = config.GetConfig()

const (
	logFolder = "logs"
	logPath   = "./logs/all.log"
)

func init() {
	if cfg.Logger.UseZap {
		startupZap()
	}
}

func startupZap() {
	var output []string

	if cfg.Logger.ToConsole {
		output = append(output, "stdout")
	}

	if cfg.Logger.ToFile {
		if _, err := os.Stat(logFolder); os.IsNotExist(err) {
			if err := os.Mkdir(logFolder, 0666); err != nil {
				panic(err)
			}
		}
		_, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("%s: %s", logPath, err)
		}
		output = append(output, logPath)
	}

	c := zap.NewProductionConfig()
	c.Development = cfg.IsDebug
	c.InitialFields = map[string]interface{}{
		"_service": cfg.Service,
	}
	c.OutputPaths = output
	logger, _ := c.Build()
	zapLogger = logger
}

// ----------------------------- Fields

type Logger struct {
	*zap.Logger
}

var (
	zapLogger       *zap.Logger
	zapLoggerEnrich *Logger
	zapOnce         sync.Once
)

func GetLogger() *Logger {
	zapOnce.Do(func() {
		zapLoggerEnrich = &Logger{zapLogger}
	})
	return zapLoggerEnrich
}
