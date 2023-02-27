package filehashservice

import (
	"context"
	"hash"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/fshasher"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/hasher"
	"github.com/sirupsen/logrus"
)

type FileHash struct {
	Path string
	Hash string
}

type FileHashService struct {
	alg     string
	path    string
	workers int
}

func New(logger *logrus.Logger, alg string, path string, workers int) *FileHashService {
	return &FileHashService{
		alg:     alg,
		path:    path,
		workers: workers,
	}
}

// Calculate calculate file hashes synchronously and store into slice
func (fhs *FileHashService) CalculateAll(ctx context.Context) ([]FileHash, error) {
	hashChan := make(chan FileHash)
	result := make([]FileHash, 0, 1024)
	var err error

	go func() {
		hashFuncBuilder := fshasher.FileHasherByHash(func() hash.Hash { return hasher.NewHashSum(fhs.alg) })
		err = fshasher.Walk(ctx, fhs.workers, fhs.path, hashFuncBuilder, func(path, hash string) error { return nil })
		close(hashChan)
	}()

	for h := range hashChan {
		result = append(result, h)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// CalculateInCallback calculate file hashes and call callback for each hash
func (fhs *FileHashService) CalculateInCallback(ctx context.Context, handlert func(fh FileHash) error) error {
	hashFuncBuilder := fshasher.FileHasherByHash(func() hash.Hash { return hasher.NewHashSum(fhs.alg) })
	return fshasher.Walk(ctx, fhs.workers, fhs.path, hashFuncBuilder, func(path, hash string) error {
		return handlert(FileHash{
			Path: path,
			Hash: hash,
		})
	})
}

// CalculateInChan calculate file hashes and send into chan
// both result and error channels will be closed at the end
func (fhs *FileHashService) CalculateInChan(ctx context.Context) (chan FileHash, chan error) {
	resultChan := make(chan FileHash)
	errChan := make(chan error)

	go func() {
		hashFuncBuilder := fshasher.FileHasherByHash(func() hash.Hash { return hasher.NewHashSum(fhs.alg) })
		err := fshasher.Walk(ctx, fhs.workers, fhs.path, hashFuncBuilder, func(path, hash string) error {
			resultChan <- FileHash{
				Path: path,
				Hash: hash,
			}
			return nil
		})
		if err != nil {
			errChan <- err
		}
		close(resultChan)
		close(errChan)
	}()
	return resultChan, errChan
}
