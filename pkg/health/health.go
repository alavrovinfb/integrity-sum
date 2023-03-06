package health

import (
	"os"
)

// Health sets either healthy or unhealthy status
type Health interface {
	Set() error
	Reset() error
}

// New returns a Health status
// Indicate whether the app is ready.
// Pod manifest should include liveness probe of "exec" kind
func New(file string) Health {
	return healthFile{
		file: file,
	}
}

type healthFile struct {
	file string
}

// Set healthy status by creating corresponding file
func (r healthFile) Set() error {
	f, err := os.Create(r.file)
	if err != nil {
		return err
	}
	return f.Close()
}

// Unset removes the healthy file.
func (r healthFile) Reset() error {
	return os.Remove(r.file)
}
