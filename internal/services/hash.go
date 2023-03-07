package services

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/hasher"
)

type HashService struct {
	alg    string
	logger *logrus.Logger
}

// NewHashService creates a new struct HashService
func NewHashService(
	alg string,
	logger *logrus.Logger) *HashService {
	return &HashService{
		alg:    alg,
		logger: logger,
	}
}

// WorkerPool launches a certain number of workers for concurrent processing
func (hs HashService) WorkerPool(jobs chan string, results chan *api.HashData) {
	countWorkers := viper.GetInt("count-workers")

	var wg sync.WaitGroup
	for w := 1; w <= countWorkers; w++ {
		wg.Add(1)
		go hs.Worker(&wg, jobs, results)
	}
	defer close(results)
	wg.Wait()
}

// Worker gets jobs from a pipe and writes the result to stdout and database
func (hs HashService) Worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- *api.HashData) {
	defer wg.Done()
	for j := range jobs {
		data, err := hs.CreateHash(j)
		if err != nil {
			hs.logger.Errorf("error creating file hash - %s, %s", j, err)
			continue
		}
		results <- data
	}
}

// CreateHash creates a new object with a hash sum
func (hs HashService) CreateHash(path string) (*api.HashData, error) {
	file, err := os.Open(path)
	if err != nil {
		hs.logger.Errorf("can not open file %s %s", path, err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			hs.logger.Errorf("[HashService]Error closing file: %s", err)
		}
	}(file)

	h := hasher.NewHashSum(hs.alg)
	_, err = io.Copy(h, file)
	if err != nil {
		return nil, err
	}
	res := hex.EncodeToString(h.Sum(nil))
	outputHashSum := api.HashData{
		Hash:         res,
		FullFileName: path,
		Algorithm:    hs.alg,
	}
	return &outputHashSum, nil
}

// IsDataChanged checks if the current data has changed with the data stored in the database
func (hs HashService) IsDataChanged(currentHashData []*api.HashData, hashDataFromDB []*models.HashData, dataFromDBbyRelease *models.Release, deploymentData *models.DeploymentData) bool {
	isDataChanged := wasDataChanged(hashDataFromDB, currentHashData, deploymentData, dataFromDBbyRelease)
	isAddedFiles := wasAddedFiles(currentHashData, hashDataFromDB)

	if isDataChanged || isAddedFiles {
		return true
	}
	return false
}

func wasDataChanged(
	hashSumFromDB []*models.HashData,
	currentHashData []*api.HashData,
	deploymentData *models.DeploymentData,
	dataFromDBbyRelease *models.Release,
) bool {
	for _, dataFromDB := range hashSumFromDB {
		trigger := false
		for _, dataCurrent := range currentHashData {
			if strings.TrimSpace(dataFromDB.FullFileName) == strings.TrimSpace(dataCurrent.FullFileName) && strings.TrimSpace(dataFromDB.Algorithm) == strings.TrimSpace(dataCurrent.Algorithm) {
				if dataFromDB.Hash != dataCurrent.Hash {
					fmt.Printf("Changed: file - %s, old hash sum %s, new hash sum %s\n",
						dataFromDB.FullFileName, dataFromDB.Hash, dataCurrent.Hash)
					return true
				}
				if dataFromDBbyRelease.Image != deploymentData.Image && dataFromDBbyRelease.Name == deploymentData.NameDeployment {
					fmt.Printf("Changed image container: file - %s, old image %s, new image %s\n",
						dataFromDB.FullFileName, dataFromDBbyRelease.Image, deploymentData.Image)
					return true
				}
				trigger = true
				break
			}
		}
		if !trigger {
			fmt.Printf("Deleted: file - %s, hash sum %s\n", dataFromDB.FullFileName, dataFromDB.Hash)
			return true
		}
	}
	return false
}

func wasAddedFiles(currentHashData []*api.HashData, hashDataFromDB []*models.HashData) bool {
	dataFromDB := make(map[string]struct{}, len(hashDataFromDB))
	for _, value := range hashDataFromDB {
		dataFromDB[value.FullFileName] = struct{}{}
	}

	for _, dataCurrent := range currentHashData {
		if _, ok := dataFromDB[dataCurrent.FullFileName]; !ok {
			fmt.Printf("Changed: the current data is different from the data in the database, current file - %s, hash sum %s\n", dataCurrent.FullFileName, dataCurrent.Hash)
			return true
		}
	}
	return false
}
