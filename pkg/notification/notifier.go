package notification

import "time"

type Message struct {
	Time    time.Time
	Message string
}

type Notifier interface {
	Send(msg Message) error
}
