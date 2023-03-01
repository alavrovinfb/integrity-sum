package integritymonitor

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/ports"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehash"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/process"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
	"github.com/sirupsen/logrus"
)

var ErrIntegrityNewFileFoud = errors.New("file found")
var ErrIntegrityFileDeleted = errors.New("file deleted")
var ErrIntegrityFileMismatch = errors.New("file content mismatch")

type IntegrityMonitor struct {
	logger             *logrus.Logger
	fshasher           *filehash.FileSystemHasher
	repository         ports.IAppRepository
	kubeclient         ports.IKuberService
	alertSender        alerts.Sender
	delay              time.Duration
	monitorProcess     string
	monitorProcessPath string
	algorithm          string
}

func New(logger *logrus.Logger,
	fshasher *filehash.FileSystemHasher,
	repository ports.IAppRepository,
	kubeclient ports.IKuberService,
	alertSender alerts.Sender,
	delay time.Duration,
	monitorProcess string,
	monitorProcessPath string,
	algorithm string,
) *IntegrityMonitor {
	return &IntegrityMonitor{
		logger:             logger,
		fshasher:           fshasher,
		repository:         repository,
		kubeclient:         kubeclient,
		alertSender:        alertSender,
		delay:              delay,
		monitorProcess:     monitorProcess,
		monitorProcessPath: monitorProcessPath,
		algorithm:          algorithm,
	}
}

func (m *IntegrityMonitor) Run(ctx context.Context) error {
	processPath, err := m.getProcessPath(m.monitorProcess, m.monitorProcessPath)
	if err != nil {
		return err
	}

	kuberData, err := m.kubeclient.ConnectionToK8sAPI()
	if err != nil {
		return err
	}

	deploymentData, err := m.kubeclient.GetDataFromDeployment(kuberData)
	if err != nil {
		return err
	}

	err = m.setupIntegrity(ctx, processPath, kuberData, deploymentData)
	if err != nil {
		m.logger.WithError(err).Error("failed setup integrity")
		return err
	}

	ticker := time.NewTicker(m.delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := m.checkIntegrity(ctx, processPath, kuberData, deploymentData)
			if err != nil {
				m.logger.WithError(err).Error("failed check integrity")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *IntegrityMonitor) setupIntegrity(ctx context.Context, directory string, kubeData *models.KuberData, deploymentData *models.DeploymentData) error {
	m.logger.Debug("begin setup integrity")
	m.logger.Trace("begin calculate hashes")
	fileHashes, err := m.fshasher.CalculateAll(ctx, directory)
	if err != nil {
		return fmt.Errorf("failed calculate file hashes: %w", err)
	}
	m.logger.WithField("filesCount", len(fileHashes)).Trace("end calculate hashes")

	fileHashesDto := make([]*api.HashData, 0, len(fileHashes))
	for _, fh := range fileHashes {
		fileHashesDto = append(fileHashesDto, &api.HashData{
			Hash:         fh.Hash,
			FileName:     path.Base(fh.Path),
			FullFilePath: fh.Path,
			Algorithm:    m.algorithm,
		})
	}

	m.logger.Trace("begin store integrity hashes into storage")
	err = m.repository.SaveHashData(fileHashesDto, deploymentData)
	if err != nil {
		return err
	}

	m.logger.Debug("end setup integrity")
	return nil
}

func (m *IntegrityMonitor) checkIntegrity(ctx context.Context, directory string, kubeData *models.KuberData, deploymentData *models.DeploymentData) error {
	fileHashes, err := m.fshasher.CalculateAll(ctx, directory)
	if err != nil {
		return fmt.Errorf("failed calculate file hashes: %w", err)
	}

	fileHashesDto, err := m.repository.GetHashData(directory, m.algorithm, deploymentData)
	if err != nil {
		return fmt.Errorf("failed get hash data: %w", err)
	}

	referecenHashes := make(map[string]*models.HashDataFromDB, len(fileHashesDto))

	for _, fh := range fileHashesDto {
		referecenHashes[fh.FullFilePath] = fh
	}

	for _, fh := range fileHashes {
		if fhdto, ok := referecenHashes[fh.Path]; ok {
			if fhdto.Hash != fh.Hash {
				m.integrityCheckFailed(ErrIntegrityFileMismatch, fh.Path, kubeData, deploymentData)
				return ErrIntegrityFileMismatch
			}
			delete(referecenHashes, fh.Path)
		} else {
			m.integrityCheckFailed(ErrIntegrityNewFileFoud, fh.Path, kubeData, deploymentData)
			return ErrIntegrityNewFileFoud
		}
	}

	if len(referecenHashes) > 0 {
		for path := range referecenHashes {
			m.integrityCheckFailed(ErrIntegrityFileDeleted, path, kubeData, deploymentData)
			return ErrIntegrityFileDeleted
		}
	}

	return err
}

func (m *IntegrityMonitor) getProcessPath(procName string, path string) (string, error) {
	pid, err := process.GetPID(procName)
	if err != nil {
		return "", fmt.Errorf("failed build process path: %w", err)
	}
	return fmt.Sprintf("/proc/%d/root/%s", pid, path), nil
}

func (m *IntegrityMonitor) integrityCheckFailed(err error, path string, kubeData *models.KuberData, deploymentData *models.DeploymentData) {
	switch err {
	case ErrIntegrityFileMismatch:
		m.logger.WithField("path", path).Warn("file content missmatch")
	case ErrIntegrityNewFileFoud:
		m.logger.WithField("path", path).Warn("new file found")
	case ErrIntegrityFileDeleted:
		m.logger.WithField("path", path).Warn("file deleted")
	}
	if m.alertSender != nil {
		err := m.alertSender.Send(alerts.Alert{
			Time:    time.Now(),
			Message: fmt.Sprintf("Restart deployment %v", deploymentData.NameDeployment),
			Reason:  "mismatch file content",
			Path:    path,
		})
		if err != nil {
			m.logger.WithError(err).Error("Failed send alert")
		}
	}
	m.kubeclient.RolloutDeployment(kubeData)
}
