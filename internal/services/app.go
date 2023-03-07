package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/models"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/ports"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
)

type AppService struct {
	ports.IHashService
	ports.IAppRepository
	ports.IKuberService
	ports.IReleaseStorageService
	ports.IHashStorageService
	alertSender alerts.Sender
	logger      *logrus.Logger
}

// NewAppService creates a new struct AppService
func NewAppService(
	r *repositories.AppRepository,
	alertSender alerts.Sender,
	serviceReleaseStorage *ReleaseStorageService,
	serviceHashStorage *HashStorageService,
	algorithm string,
	logger *logrus.Logger,
) *AppService {
	return &AppService{
		IHashService:           NewHashService(strings.ToUpper(algorithm), logger),
		IAppRepository:         r,
		IKuberService:          NewKuberService(logger),
		IReleaseStorageService: serviceReleaseStorage,
		IHashStorageService:    serviceHashStorage,
		alertSender:            alertSender,
		logger:                 logger,
	}
}

// GetPID returns process PID by name
func (as *AppService) GetPID(procName string) (pid int, err error) {
	cmdOut, err := exec.Command("pidof", procName).Output()
	if err != nil {
		as.logger.WithField("procName", procName).WithError(err).Error("GetPID(): proc name not found")
		return
	}
	// if found, we always have a list of PIDs
	ss := strings.Split(string(cmdOut), " ")
	pid, err = strconv.Atoi(strings.TrimSpace(ss[0]))
	return
}

// LaunchHasher takes a path to a directory and returns HashData
func (as *AppService) LaunchHasher(ctx context.Context, dirPath string) []*api.HashData {
	jobs := make(chan string)
	results := make(chan *api.HashData)
	go as.IHashService.WorkerPool(jobs, results)
	go api.SearchFilePath(ctx, dirPath, jobs, as.logger)
	allHashData := api.Result(ctx, results)

	return allHashData
}

// IsExistDeploymentNameInDB checks if the database is empty
func (as *AppService) IsExistDeploymentNameInDB(deploymentName string) bool {
	isEmptyDB, err := as.IAppRepository.IsExistDeploymentNameInDB(deploymentName)
	if err != nil && err != sql.ErrNoRows {
		as.logger.Fatalf("database check error %s", err)
	}
	return isEmptyDB
}

// Start getting the hash sum of all files, outputs to os.Stdout and saves to the database
func (as *AppService) Start(ctx context.Context, dirPath string, deploymentData *models.DeploymentData) error {
	allHashData := as.LaunchHasher(ctx, dirPath)
	err := as.IReleaseStorageService.Create(deploymentData)
	if err != nil {
		as.logger.Error("Error save hash data to database ", err)
		return err
	}

	err = as.IHashStorageService.Create(allHashData, deploymentData)
	if err != nil {
		as.logger.Error("Error save hash data to database ", err)
		return err
	}
	return nil
}

// Check getting the hash sum of all files, matches them and outputs to os.Stdout changes
func (as *AppService) Check(ctx context.Context, dirPath string, deploymentData *models.DeploymentData, kuberData *models.KuberData) error {
	hashDataCurrentByDirPath := as.LaunchHasher(ctx, dirPath)

	dataFromDBbyPodName, err := as.IHashStorageService.Get(dirPath, deploymentData)
	if err != nil {
		as.logger.Error("Error getting hash data from database ", err)
		return err
	}
	dataFromDBbyRelease, err := as.IReleaseStorageService.Get(deploymentData)
	if err != nil {
		as.logger.Error("Error getting hash data from database ", err)
		return err
	}

	isDataChanged := as.IHashService.IsDataChanged(hashDataCurrentByDirPath, dataFromDBbyPodName, dataFromDBbyRelease, deploymentData)
	if isDataChanged {

		// alert sender is optional
		if as.alertSender != nil {
			err := as.alertSender.Send(alerts.Alert{
				Time:    time.Now(),
				Message: fmt.Sprintf("Restart deployment %v", deploymentData.NameDeployment),
				Reason:  "mismatch file content",
				Path:    dirPath,
			})
			if err != nil {
				as.logger.WithError(err).Error("Failed send alert")
			}
		}

		err = as.IReleaseStorageService.Delete(deploymentData.NameDeployment)
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
