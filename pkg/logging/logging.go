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
	"sync"
)

var cfg = config.GetConfig()

func init() {
	l := logrus.New()
	l.SetReportCaller(true)
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

	// Log to file
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		if err := os.Mkdir("logs", 0644); err != nil {
			panic(err)
		}
	}
	file, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		log.Fatalf("logs/all.log: %s", err)
	}

	// Log to Graylog
	gelfWriter, err := gelf.NewUDPWriter(cfg.GraylogAddr)
	if err != nil {
		log.Fatalf("gelf.NewWriter: %s", err)
	}

	l.SetOutput(io.Discard)
	l.AddHook(&writerHook{
		Writer:    []io.Writer{file, os.Stdout, gelfWriter},
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

//func GetLoggerWithFields(fields logrus.Fields) *Logger {
//	once.Do(func() {
//		l = &Logger{e.
//			WithField("_service", cfg.Service).
//			WithFields(fields)}
//	})
//	return l
//}
