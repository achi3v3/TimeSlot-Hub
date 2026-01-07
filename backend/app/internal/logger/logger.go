package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger обертка над logrus.Logger
type Logger struct {
	*logrus.Logger
}

// NewLogger создает новый экземпляр логгера
func NewLogger() *Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return &Logger{Logger: log}
}
