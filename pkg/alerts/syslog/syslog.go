package syslog

import (
	"fmt"
	"log"
	"log/syslog"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
)

const (
	DefaultPriority = syslog.LOG_WARNING | syslog.LOG_DAEMON
	fileMismatch    = 1
)

var _ alerts.Sender = (*SyslogClient)(nil)

type SyslogClient struct {
	logger    *logrus.Logger
	writer    *syslog.Writer
	bytesSent int64
}

// New creates syslog client
// priority syslog.LOG_WARNING|syslog.LOG_DAEMON
func New(logger *logrus.Logger, network, addr string, priority syslog.Priority) (*SyslogClient, error) {
	sysLog, err := syslog.Dial(network, addr, priority, alerts.AppId)
	if err != nil {
		log.Fatal(err)
	}
	return &SyslogClient{
		logger: logger,
		writer: sysLog,
	}, nil
}

func (sl *SyslogClient) Send(alert alerts.Alert) error {
	n, err := fmt.Fprintf(sl.writer, sl.syslogMessage(alert))
	sl.bytesSent += int64(n)
	return err
}

func (sl *SyslogClient) Close() {
	sl.Close()
}

func (sl *SyslogClient) syslogMessage(alert alerts.Alert) string {
	return fmt.Sprintf("time=%s event-type=%04d service=%s namespace=%s cluster=%s message=%s file=%s reason=%s",
		alert.Time, fileMismatch, viper.GetString("process"), viper.GetString("pod-namespace"),
		viper.GetString("cluster-name"), alert.Message, alert.Path, alert.Reason)
}
