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

	dirName, _ := filepath.Abs("./")
	fileC := ChanWalkDir(ctx, dirName)

	for v := range fileC {
		logrus.Infof("file: %s", v)
		cancel()
	}
}
