package logger

import "github.com/sirupsen/logrus"

func New() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.InfoLevel)
	l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	return l
}
