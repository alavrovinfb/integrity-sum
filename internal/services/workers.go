package services

import (
	"sync"
)

type HashWorker func(ind int, fileNames <-chan string, hashes chan<- string)

func WorkersPool(fileNames <-chan string, w HashWorker) <-chan string {
	countWorkers := 2 // TODO: fix
	hashChan := make(chan string, countWorkers)
	go func() {
		var wg sync.WaitGroup
		wg.Add(countWorkers)
		for i := 0; i < countWorkers; i++ {
			go func(ind int, wg *sync.WaitGroup) {
				defer wg.Done()
				w(ind, fileNames, hashChan)
			}(i, &wg)
		}
		wg.Wait()
		close(hashChan)
	}()

	return hashChan
}
