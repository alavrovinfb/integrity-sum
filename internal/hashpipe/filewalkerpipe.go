package hashpipe

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func WalkFiles(ctx context.Context, log *logrus.Logger, direcotries []string, errc chan<- error) <-chan string {
	output := make(chan string)
	go func() {
		for _, dirPath := range direcotries {
			err := filepath.WalkDir(dirPath, func(filePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// skip dir
				if d.IsDir() {
					return nil
				}

				// skip simlink
				if (d.Type() & fs.ModeSymlink) > 0 {
					return nil
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			})

			if err != nil {
				if ctx.Err() != err {
					errc <- err
				}
				break
			}
		}
		close(output)
	}()
	return output
}
