package logger

import "github.com/sirupsen/logrus"

func Init(ll string) *logrus.Logger {
	l := logrus.New()
	l.SetReportCaller(true)

	lvl := logrus.InfoLevel
	if len(ll) > 0 {
		var err error
		lvl, err = logrus.ParseLevel(ll)
		if err != nil {
			l.WithError(err).Error("Failed parse log level")
		}
	}
	l.Level = lvl
	return l
}
