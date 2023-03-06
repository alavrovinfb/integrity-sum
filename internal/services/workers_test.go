package services

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestWorkersPool(t *testing.T) {
	namesChannel := make(chan string)
	hashesChannel := WorkersPool(namesChannel, worker)
	var testData = map[string]struct{}{
		"name1": {},
		"name2": {},
		"name3": {},
		"name4": {},
		"name5": {},
	}

	go func() {
		for k := range testData {
			namesChannel <- k
		}
		close(namesChannel)
	}()

	cnt := 0
	for v := range hashesChannel {
		logrus.WithField("hash", v).Info("got hash")
		_, ok := testData[v]
		assert.True(t, ok)
		cnt++
	}
	assert.Equal(t, len(testData), cnt)
}

func worker(ind int, fileNames <-chan string, hashes chan<- string) {
	logrus.WithField("ind", ind).Info("worker started")
	for v := range fileNames {
		hashes <- v
	}
	logrus.WithField("ind", ind).Info("worker stopped")
}
