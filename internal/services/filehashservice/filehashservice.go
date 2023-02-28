package filehashservice

import (
	"context"
	"hash"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/model"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/fshasher"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/hasher"
	"github.com/sirupsen/logrus"
)

type FileHashService struct {
	log     *logrus.Logger
	alg     string
	path    string
	workers int
}

func New(log *logrus.Logger, alg string, path string, workers int) *FileHashService {
	return &FileHashService{
		log:     log,
		alg:     alg,
		path:    path,
		workers: workers,
	}
}

// Calculate calculate file hashes synchronously and store into slice
func (fhs *FileHashService) CalculateAll(ctx context.Context) ([]model.FileHash, error) {
	hashChan := make(chan model.FileHash)
	result := make([]model.FileHash, 0, 1024)
	var err error

	go func() {
		hashFuncBuilder := fshasher.FileHasherByHash(func() hash.Hash { return hasher.NewHashSum(fhs.alg) })
		fhs.log.Debug("Begin calculate hashes")
		err = fshasher.Walk(ctx, fhs.workers, fhs.path, hashFuncBuilder, func(path, hash string) error { return nil })
		close(hashChan)
	}()

	for h := range hashChan {
		result = append(result, h)
	}

	if err != nil {
		fhs.log.WithError(err).Debug("Failed calculate hashes")
		return nil, err
	}

	fhs.log.WithField("HashNum", len(result)).Debug("Success calculate hashes")

	return result, nil
}

// CalculateInCallback calculate file hashes and call callback for each hash
func (fhs *FileHashService) CalculateInCallback(ctx context.Context, handlert func(fh model.FileHash) error) error {
	hashFuncBuilder := fshasher.FileHasherByHash(func() hash.Hash { return hasher.NewHashSum(fhs.alg) })
	fhs.log.Debug("Begin calculate hashes")
	err := fshasher.Walk(ctx, fhs.workers, fhs.path, hashFuncBuilder, func(path, hash string) error {
		fhs.log.Tracef("Hash calculated : %v", path)
		return handlert(model.FileHash{
			Path: path,
			Hash: hash,
		})
	})
	if err != nil {
		fhs.log.WithError(err).Debug("Failed calculate hashes")
		return err
	}
	fhs.log.Debug("Success calculate hashes")
	return nil
}

// CalculateInChan calculate file hashes and send into chan
// both result and error channels will be closed at the end
func (fhs *FileHashService) CalculateInChan(ctx context.Context) (chan model.FileHash, chan error) {
	resultChan := make(chan model.FileHash)
	errChan := make(chan error)

	go func() {
		hashFuncBuilder := fshasher.FileHasherByHash(func() hash.Hash { return hasher.NewHashSum(fhs.alg) })
		fhs.log.Debug("Begin calculate hashes")
		err := fshasher.Walk(ctx, fhs.workers, fhs.path, hashFuncBuilder, func(path, hash string) error {
			fhs.log.Tracef("Hash calculated : %v", path)
			resultChan <- model.FileHash{
				Path: path,
				Hash: hash,
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
