package fshasher

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/ScienceSoft-Inc/integrity-sum/pkg/hasher"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type HashProcessor func(filePath string, fhash string) error

// Walk is walking through directory and subdirectories, calculate hashes of files and call processor for hash results
// error stop execution
func Walk(ctx context.Context, log *logrus.Logger, workers int, dirPath string, alg string, processor HashProcessor) error {
	if workers <= 0 {
		workers = 1
	}
	filesChan := make(chan string, 1024)
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(-1)

	group.Go(func() error {
		err := walkDir(groupCtx, dirPath, filesChan)
		close(filesChan)
		return err
	})

	for i := 0; i < workers; i++ {
		group.Go(func() error {
			fileHasher := hasher.NewFileHasher(alg, log)
			for {
				select {
				case <-groupCtx.Done():
					return nil
				case filePath, ok := <-filesChan:
					if !ok {
						return nil
					}
					hash, err := fileHasher.HashFile(filePath)
					if err != nil {
						return err
					}
					err = processor(filePath, hash)
					if err != nil {
						return err
					}
				}
			}
		})
	}
	return group.Wait()
}

func walkDir(ctx context.Context, dirPath string, outputChan chan<- string) error {
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
		case outputChan <- filePath:
			return nil
		}
	})
	return err
}
