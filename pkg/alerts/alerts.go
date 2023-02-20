package alerts

import "time"

type Alert struct {
	Time    time.Time
	Message string
	Reason  string
	Path    string
}

type Sender interface {
	Send(alert Alert) error
}
