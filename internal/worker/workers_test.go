package worker

import (
	"testing"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehash"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestWorkersPool(t *testing.T) {
	namesChannel := make(chan string)
	hashesChannel := WorkersPool(3, namesChannel, mockWorker)
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
		_, ok := testData[v.Path]
		assert.True(t, ok)
		cnt++
	}
	assert.Equal(t, len(testData), cnt)
}

func mockWorker(ind int, fileNameC <-chan string, hashC chan<- filehash.FileHash) {
	logrus.WithField("ind", ind).Info("worker started")
	for v := range fileNameC {
		hashC <- filehash.FileHash{
			Path: v,
			Hash: v,
		}
	}
	logrus.WithField("ind", ind).Info("worker stopped")
}
