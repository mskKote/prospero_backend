package logging

import (
	"fmt"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/mskKote/go-gelf.v2/gelf"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var cfg = config.GetConfig()

func init() {
	l := logrus.New()
	l.SetReportCaller(true)
	if cfg.Logger.IsJSON {
		l.Formatter = &logrus.JSONFormatter{
			CallerPrettyfier: func(frame *runtime.Frame) (function, file string) {
				filename := path.Base(frame.File)
				return fmt.Sprintf("%s()", frame.Function),
					fmt.Sprintf("%s:%d", filename, frame.Line)
			},
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg:         "short_message",
				logrus.FieldKeyTime:        "timestamp",
				logrus.FieldKeyLevel:       "_level",
				logrus.FieldKeyLogrusError: "_error",
				logrus.FieldKeyFunc:        "_caller",
			},
			DisableTimestamp: true,
		}
	} else {
		l.Formatter = &logrus.TextFormatter{
			ForceColors:            true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			FullTimestamp:          true,
			TimestampFormat:        time.DateTime,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := strings.Split(f.File, "/")
				filenameShort := strings.Join(filename[len(filename)-3:], "/")
				return "", fmt.Sprintf("{...%s:%d}", filenameShort, f.Line)
			},
		}
	}

	writers := []io.Writer{os.Stdout}
	// Log to file
	if cfg.Logger.ToFile {
		if _, err := os.Stat("logs"); os.IsNotExist(err) {
			if err := os.Mkdir("logs", 0644); err != nil {
				panic(err)
			}
		}
		file, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			log.Fatalf("logs/all.log: %s", err)
		}
		writers = append(writers, file)
	}

	// Log to Graylog
	if cfg.Logger.ToGraylog {
		gelfWriter, err := gelf.NewUDPWriter(cfg.Logger.GraylogAddr)
		if err != nil {
			log.Fatalf("gelf.NewWriter: %s", err)
		}
		writers = append(writers, gelfWriter)
	}

	l.SetOutput(io.Discard)
	l.AddHook(&writerHook{
		Writer:    writers,
		LogLevels: logrus.AllLevels,
	})
	l.SetLevel(logrus.TraceLevel)

	e = &logrus.Entry{Logger: l}
}

// hook to write to multiply sources
type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	for _, writer := range hook.Writer {
		_, errWriter := writer.Write([]byte(line))
		if errWriter != nil {
			return errWriter
		}
	}
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

// Custom Loggers

type Logger struct {
	*logrus.Entry
}

var (
	e    *logrus.Entry
	l    *Logger
	once sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		l = &Logger{e.WithField("_service", cfg.Service)}
	})
	return l
}

func GetLoggerWithFields(fields logrus.Fields) *Logger {
	once.Do(func() {
		l = &Logger{e.
			WithField("_service", cfg.Service).
			WithFields(fields)}
	})
	return l
}
