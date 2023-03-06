package walker

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func ChanWalkDir(ctx context.Context, dirPath string) <-chan string {
	fileNamesChan := make(chan string)
	go func() {
		defer close(fileNamesChan)

		if err := filepath.WalkDir(dirPath, func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// skip dir
			if d.IsDir() {
				return nil
			}

			// skip symlink
			if (d.Type() & fs.ModeSymlink) > 0 {
				return nil
			}

			select {
			case fileNamesChan <- filePath:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil {
			logrus.WithError(err).Error("file walker")
		}
	}()

	return fileNamesChan
}
