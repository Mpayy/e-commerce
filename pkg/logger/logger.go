package logger

import (
	"os"

	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields

type Logger struct {
	*logrus.Logger
}

type Entry struct {
	*logrus.Entry
}

func NewLogger(cfg *config.Config) *Logger {
	l := logrus.New()

	l.SetOutput(os.Stdout)

	if cfg.AppEnv == "production" {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		l.SetLevel(logrus.InfoLevel)
	} else {
		l.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
		l.SetLevel(logrus.DebugLevel)
	}

	return &Logger{Logger: l}
}

func (l *Logger) WithFields(fields Fields) *Entry {
	return &Entry{
		Entry: l.Logger.WithFields(logrus.Fields(fields)),
	}
}

func (e *Entry) WithFields(fields Fields) *Entry {
	return &Entry{
		Entry: e.Entry.WithFields(logrus.Fields(fields)),
	}
}
