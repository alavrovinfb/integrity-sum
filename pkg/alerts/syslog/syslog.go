package syslog

import (
	"fmt"
	"log/syslog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
)

const (
	DefaultPriority = syslog.LOG_WARNING | syslog.LOG_DAEMON
)

var ErrToType = map[string]int{
	"file content mismatch": 1,
	"new file found":        2,
	"file deleted":          3,
	"heartbeat event":       4,
}

var _ alerts.Sender = (*SyslogClient)(nil)

type SyslogClient struct {
	logger   *logrus.Logger
	conn     net.Conn
	priority syslog.Priority
	tag      string
	hostname string
}

// New creates syslog client
// priority syslog.LOG_WARNING|syslog.LOG_DAEMON
// hostName custom hostname if empty use host a name obtained from os.Hostname()
func New(logger *logrus.Logger, network, addr string, priority syslog.Priority, hostName, tag string) (*SyslogClient, error) {
	sysLogConn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	if hostName == "" {
		hostName, _ = os.Hostname()
	}

	if tag == "" {
		tag = os.Args[0]
	}

	return &SyslogClient{
		logger:   logger,
		conn:     sysLogConn,
		priority: priority,
		hostname: hostName,
		tag:      tag,
	}, nil
}

func (sl *SyslogClient) Send(alert alerts.Alert) error {
	msg := sl.syslogMessage(alert)
	nl := ""
	if !strings.HasSuffix(msg, "\n") {
		nl = "\n"
	}
	// Syslog record header <PRI>TIMESTAMP HOST TAG
	header := fmt.Sprintf("<%d>%s %s %s[%d]:",
		sl.priority,
		time.Now().Format(time.Stamp),
		sl.hostname,
		sl.tag,
		os.Getpid(),
	)

	_, err := fmt.Fprintf(sl.conn, "%s %s%s", header, msg, nl)

	return err
}

func (sl *SyslogClient) Close() {
	sl.conn.Close()
}

func (sl *SyslogClient) syslogMessage(alert alerts.Alert) string {
	pn := alert.ProcessName
	return fmt.Sprintf("time=%s event-type=%04d service=%s image=%s namespace=%s cluster=%s message=%s file=%s reason=%s",
		alert.Time.Format(time.Stamp), ErrToType[alert.Reason], pn, viper.GetStringMapString("process-image")[pn],
		viper.GetString("pod-namespace"), viper.GetString("cluster-name"), alert.Message, alert.Path, alert.Reason)
}
