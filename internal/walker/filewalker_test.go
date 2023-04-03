package walker

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestChanWalkDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log := logrus.New()

	dirName, _ := filepath.Abs("./")
	fileC := ChanWalkDir(ctx, []string{dirName}, log)

	for v := range fileC {
		log.Infof("file: %s", v)
		cancel()
	}
}
