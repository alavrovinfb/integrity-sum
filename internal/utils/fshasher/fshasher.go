package fshasher

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

type FileHasher func(path string) (string, error)
type HashProcessor func(path string, fhash string) error

// FileHasherByHash make FileHasher function from hash.Hash
// each hash execution require new hasher object
func FileHasherByHash(hashBuilder func() hash.Hash) FileHasher {
	return func(path string) (string, error) {
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Printf("Failed close file: %v", err)
			}
		}(file)

		h := hashBuilder()
		_, err = io.Copy(h, file)
		if err != nil {
			return "", err
		}
		res := hex.EncodeToString(h.Sum(nil))
		return res, nil
	}
}

// Walk is walking through directory and subdirectories, calculate hashes of files and call processor for hash results
// error stop execution
func Walk(ctx context.Context, workers int, path string, fileHasher FileHasher, processor HashProcessor) error {
	if workers <= 0 {
		workers = 1
	}
	filesChan := make(chan string, 1024)
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(-1)

	group.Go(func() error {
		err := walkDir(groupCtx, path, filesChan)
		close(filesChan)
		return err
	})

	for i := 0; i < workers; i++ {
		group.Go(func() error {
			for {
				select {
				case <-groupCtx.Done():
					return nil
				case path, ok := <-filesChan:
					if !ok {
						return nil
					}
					hash, err := fileHasher(path)
					if err != nil {
						return err
					}
					err = processor(path, hash)
					if err != nil {
						return err
					}
				}
			}
		})
	}
	return group.Wait()
}

func walkDir(ctx context.Context, path string, outputChan chan<- string) error {
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
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
		case outputChan <- path:
			return nil
		}
	})
	return err
}
