package walker

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func ChanWalkDir(ctx context.Context, dirPaths []string, log *logrus.Logger) <-chan string {
	fileNamesChan := make(chan string)
	go func() {
		defer close(fileNamesChan)
		for _, dirPath := range dirPaths {
			if err := filepath.WalkDir(dirPath, func(filePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !d.Type().IsRegular() {
					return nil
				}

				select {
				case fileNamesChan <- filePath:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}); err != nil {
				log.WithError(err).Error("file walker")
			}
		}
	}()

	return fileNamesChan
}
