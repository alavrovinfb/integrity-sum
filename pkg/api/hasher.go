package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// SearchFilePath searches for all files in the given directory
func SearchFilePath(commonPath string, jobs chan<- string, logger *logrus.Logger) {
	err := filepath.Walk(commonPath, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			jobs <- path
		}
		if err != nil {
			logger.Error("err while going to path files", err)
			return err
		}

		return nil
	})
	close(jobs)

	if err != nil {
		logger.Error("not exist directory path", err)
		return
	}
}

// Result launching an infinite loop of receiving and outputting to Stdout the result and signal control
func Result(ctx context.Context, results chan *HashData, c chan os.Signal) []*HashData {
	var allHashData []*HashData
	for {
		select {
		case hashData, ok := <-results:
			if !ok {
				return allHashData
			}
			allHashData = append(allHashData, hashData)
		case <-c:
			fmt.Println("exit program")
			return nil
		case <-ctx.Done():
			fmt.Println("program termination after receiving a signal")
			return nil
		}
	}
}
