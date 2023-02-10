package services

import (
	"bufio"
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/integrity-sum/internal/core/models"
	"github.com/integrity-sum/internal/core/ports"
	"github.com/integrity-sum/internal/repositories"
	"github.com/integrity-sum/pkg/api"
)

type AppService struct {
	ports.IHashService
	ports.IAppRepository
	ports.IKuberService
	logger *logrus.Logger
}

// NewAppService creates a new struct AppService
func NewAppService(r *repositories.AppRepository, algorithm string, logger *logrus.Logger) *AppService {
	algorithm = strings.ToUpper(algorithm)
	IHashService := NewHashService(r.IHashRepository, algorithm, logger)
	kuberService := NewKuberService(logger)
	return &AppService{
		IHashService:   IHashService,
		IAppRepository: r,
		IKuberService:  kuberService,
		logger:         logger,
	}
}

// GetPID getting pid by process name
func (as *AppService) GetPID(configData *models.ConfigMapData) (int, error) {
	if os.Chdir(viper.GetString("proc-dir")) != nil {
		as.logger.Error("/proc unavailable")
		return 0, errors.New("error changing the current working directory to the named directory")
	}

	files, err := os.ReadDir(".")
	if err != nil {
		as.logger.Error("unable to read /proc directory")
		return 0, err
	}
	var pid int
	for _, file := range files {
		if !file.IsDir() {
			as.logger.Info("file isn't a directory")
			return 0, err
		}

		// Our directory name should convert to integer if it's a PID
		pid, err = strconv.Atoi(file.Name())
		if err != nil {
			return 0, err
		}

		// Open the /proc/xxx/stat file to read the name
		f, err := os.Open(file.Name() + "/stat")
		if err != nil {
			as.logger.Error("unable to open", file.Name())
			return 0, err
		}
		defer f.Close()

		r := bufio.NewReader(f)
		scanner := bufio.NewScanner(r)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), configData.ProcName) {
				return pid, nil
			}
		}
	}

	return pid, nil
}

// LaunchHasher takes a path to a directory and returns HashData
func (as *AppService) LaunchHasher(ctx context.Context, dirPath string, sig chan os.Signal) []*api.HashData {
	jobs := make(chan string)
	results := make(chan *api.HashData)
	go as.IHashService.WorkerPool(jobs, results)
	go api.SearchFilePath(dirPath, jobs, as.logger)
	allHashData := api.Result(ctx, results, sig)

	return allHashData
}

// IsExistDeploymentNameInDB checks if the database is empty
func (as *AppService) IsExistDeploymentNameInDB(deploymentName string) bool {
	isEmptyDB, err := as.IAppRepository.IsExistDeploymentNameInDB(deploymentName)
	if err != nil {
		as.logger.Fatalf("database check error %s", err)
	}
	return isEmptyDB
}

// Start getting the hash sum of all files, outputs to os.Stdout and saves to the database
func (as *AppService) Start(ctx context.Context, dirPath string, sig chan os.Signal, deploymentData *models.DeploymentData) error {
	allHashData := as.LaunchHasher(ctx, dirPath, sig)
	err := as.IHashService.SaveHashData(allHashData, deploymentData)
	if err != nil {
		as.logger.Error("Error save hash data to database ", err)
		return err
	}

	return nil
}

// Check getting the hash sum of all files, matches them and outputs to os.Stdout changes
func (as *AppService) Check(ctx context.Context, dirPath string, sig chan os.Signal, deploymentData *models.DeploymentData, kuberData *models.KuberData) error {
	hashDataCurrentByDirPath := as.LaunchHasher(ctx, dirPath, sig)

	dataFromDBbyPodName, err := as.IHashService.GetHashData(dirPath, deploymentData)
	if err != nil {
		as.logger.Error("Error getting hash data from database ", err)
		return err
	}

	isDataChanged := as.IHashService.IsDataChanged(hashDataCurrentByDirPath, dataFromDBbyPodName, deploymentData)
	if isDataChanged {
		err := as.IHashService.DeleteFromTable(deploymentData.NameDeployment)
		if err != nil {
			as.logger.Error("Error while deleting rows in database", err)
			return err
		}

		err = as.IKuberService.RolloutDeployment(kuberData)
		if err != nil {
			as.logger.Error("Error while rolling out deployment in k8s", err)
			return err
		}
	}
	return nil
}
