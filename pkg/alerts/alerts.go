package alerts

import (
	"fmt"
	"time"
)

const (
	AppId      = "integrity-monitor"
	TypeSplunk = "splunk"
	TypeSyslog = "syslog"
)

type Alert struct {
	Time    time.Time
	Message string
	Reason  string
	Path    string
}

type Registry map[string]Sender

type Errors []error

type Sender interface {
	Send(alert Alert) error
}

func (r Registry) Add(senderType string, s Sender) {
	r[senderType] = s
}

func (r Registry) Send(alert Alert) error {
	var errs *Errors
	for _, s := range r {
		if err := s.Send(alert); err != nil {
			errs.Collect(err)
		}
	}

	return errs
}

func (es *Errors) Collect(e error) { *es = append(*es, e) }

func (es *Errors) Error() (err string) {
	err = "alert errors: "
	for i, e := range *es {
		err += fmt.Sprintf("Error %d: %s\n", i, e.Error())
	}

	return err
}
