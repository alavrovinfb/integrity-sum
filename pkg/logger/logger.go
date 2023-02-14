package logger

import (
	"github.com/sirupsen/logrus"
)

func InitLogger(ll int) *logrus.Logger {
	l := logrus.New()
	l.Level = logrus.Level(ll)
	l.SetReportCaller(true)

	return l
}
