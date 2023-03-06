package hashpipe

import (
	"context"

	"github.com/ScienceSoft-Inc/integrity-sum/pkg/hasher"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type FileHash struct {
	Path string
	Alg  string
	Hash string
}

func FileHashPipe(ctx context.Context, log *logrus.Logger, workers int, alg string, filePathChan <-chan string, errChan chan<- error) <-chan FileHash {
	outputChan := make(chan FileHash)
	go func() {
		group, _ := errgroup.WithContext(ctx)
		group.Go(func() error {
			fh := hasher.NewFileHasher(alg, log)

			for filePath := range filePathChan {
				hash, err := fh.HashFile(filePath)
				if err != nil {
					return err
				}
				outputChan <- FileHash{
					Path: filePath,
					Alg:  alg,
					Hash: hash,
				}
			}
			return nil
		})

		err := group.Wait()
		if err != nil && err != ctx.Err() {
			errChan <- err
		}
		close(outputChan)
	}()
	return outputChan
}
