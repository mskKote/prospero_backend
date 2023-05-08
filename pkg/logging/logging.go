package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"runtime"
	"sync"
)

func init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			filename := path.Base(frame.File)
			return fmt.Sprintf("%s()", frame.Function),
				fmt.Sprintf("%s:%d", filename, frame.Line)
		},
		DisableColors: true,
		FullTimestamp: true,
	}

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		if err := os.Mkdir("logs", 0644); err != nil {
			panic(err)
		}
	}

	file, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		panic(err)
	}

	l.SetOutput(io.Discard)
	l.AddHook(&writerHook{
		Writer:    []io.Writer{file, os.Stdout},
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
		l = &Logger{e}
	})
	return l
}

// EXAMPLE
//func GetLoggerWithField(k string, v interface{}) *Logger {
//	return &Logger{e.WithField(k, v)}
//}
