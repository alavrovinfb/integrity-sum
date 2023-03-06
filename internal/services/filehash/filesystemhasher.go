package filehash

import (
	"context"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/fshasher"
	"github.com/sirupsen/logrus"
)

type FileHash struct {
	Path string
	Hash string
}

type FileSystemHasher struct {
	log     *logrus.Logger
	alg     string
	workers int
}

func NewFileSystemHasher(log *logrus.Logger, alg string, workers int) *FileSystemHasher {
	return &FileSystemHasher{
		log:     log,
		alg:     alg,
		workers: workers,
	}
}

// Calculate calculate file hashes synchronously and store into slice
func (fhs *FileSystemHasher) CalculateAll(ctx context.Context, dirPath string) ([]FileHash, error) {
	hashChan := make(chan FileHash)
	result := make([]FileHash, 0, 1024)
	var err error

	go func() {
		fhs.log.Debug("Begin calculate hashes")
		err = fshasher.Walk(ctx, fhs.log, fhs.workers, dirPath, fhs.alg, func(filePath string, fileHash string) error {
			hashChan <- FileHash{Path: filePath, Hash: fileHash}
			return nil
		})
		close(hashChan)
	}()

	for h := range hashChan {
		result = append(result, h)
	}

	if err != nil {
		fhs.log.WithError(err).Debug("Failed calculate hashes")
		return nil, err
	}

	fhs.log.WithField("HashesCount", len(result)).Debug("Success calculate hashes")

	return result, nil
}

// CalculateInChan calculate file hashes and send into chan
// both result and error channels will be closed at the end
func (fhs *FileSystemHasher) CalculateInChan(ctx context.Context, dirPath string) (chan FileHash, chan error) {
	resultChan := make(chan FileHash)
	errChan := make(chan error)

	go func() {
		fhs.log.Debug("Begin calculate hashes")
		err := fshasher.Walk(ctx, fhs.log, fhs.workers, dirPath, fhs.alg, func(filePath, fileHash string) error {
			fhs.log.Tracef("Hash calculated : %v", filePath)
			resultChan <- FileHash{
				Path: filePath,
				Hash: fileHash,
			}
			return nil
		})
		if err != nil {
			fhs.log.WithError(err).Debug("Failed calculate hashes")
			errChan <- err
		} else {
			fhs.log.Debug("Success calculate hashes")
		}
		close(resultChan)
		close(errChan)
	}()
	return resultChan, errChan
}
