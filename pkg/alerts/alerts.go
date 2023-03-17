package alerts

import (
	"time"
)

type Alert struct {
	Time    time.Time
	Message string
	Reason  string
	Path    string
}

func New(msg, reason, path string) Alert {
	return Alert{
		Time:    time.Now(),
		Message: msg,
		Reason:  reason,
		Path:    path,
	}
}

type Sender interface {
	Send(alert Alert) error
}

var registry = []Sender{}

func Register(s Sender) {
	registry = append(registry, s)
}

func Send(alert Alert) error {
	var errs Errors
	for _, s := range registry {
		if err := s.Send(alert); err != nil {
			errs.collect(err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
