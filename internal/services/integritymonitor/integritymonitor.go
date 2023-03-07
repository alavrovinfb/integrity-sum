package integritymonitor

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/ports"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/services"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories/data"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehash"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/process"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/walker"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/worker"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
)

var ErrIntegrityNewFileFoud = errors.New("new file found")
var ErrIntegrityFileDeleted = errors.New("file deleted")
var ErrIntegrityFileMismatch = errors.New("file content mismatch")

type IntegrityMonitor struct {
	logger              *logrus.Logger
	fshasher            *filehash.FileSystemHasher
	repository          ports.IAppRepository
	alertSender         alerts.Sender
	delay               time.Duration
	algorithm           string
	monitoringDirectory string
}

func New(logger *logrus.Logger,
	fshasher *filehash.FileSystemHasher,
	repository ports.IAppRepository,
	alertSender alerts.Sender,
	delay time.Duration,
	monitorProcess string,
	monitorProcessPath string,
	algorithm string,
) (*IntegrityMonitor, error) {
	// TODO: upper layer
	processPath, err := GetProcessPath(monitorProcess, monitorProcessPath)
	if err != nil {
		return nil, err
	}

	return &IntegrityMonitor{
		logger:              logger,
		fshasher:            fshasher,
		repository:          repository,
		alertSender:         alertSender,
		delay:               delay, // TODO: remove
		algorithm:           algorithm,
		monitoringDirectory: processPath,
	}, nil
}

func (m *IntegrityMonitor) Run(ctx context.Context) error {
	// TODO: viper.GetDuration("duration-time")
	ticker := time.NewTicker(m.delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := m.checkIntegrity(ctx)
			if err != nil {
				m.logger.WithError(err).Error("failed check integrity")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func SetupIntegrity(ctx context.Context, monitoringDirectory string, log *logrus.Logger) error {
	log.Debug("begin setup integrity")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errC := make(chan error)
	defer close(errC)

	dataK8s, err := services.NewKuberService(log).GetDataFromK8sAPI()
	if err != nil {
		return err
	}

	log.Trace("calculate & save hashes...")
	for {
		select {
		case <-ctx.Done():
			log.Error(ctx.Err())
			return ctx.Err()

		case countHashes := <-saveHashes(
			ctx,
			worker.WorkersPool(
				viper.GetInt("count-workers"),
				walker.ChanWalkDir(ctx, monitoringDirectory, log),
				worker.NewWorker(ctx, viper.GetString("algorithm"), log),
			),
			dataK8s.DeploymentData,
			errC,
		):
			log.WithField("countHashes", countHashes).Info("hashes stored successfully")
			log.Debug("end setup integrity")
			return nil

		case err := <-errC:
			log.WithError(err).Error("setup integrity failed")
			return err
		}
	}
}

func (m *IntegrityMonitor) checkIntegrity(ctx context.Context) error {
	m.logger.Debug("begin check integrity")
	fileHashes, err := m.fshasher.CalculateAll(ctx, m.monitoringDirectory)
	if err != nil {
		m.logger.WithError(err).Error("failed calculate file hashes")
		return err
	}

	k8sData, err := services.NewKuberService(m.logger).GetDataFromK8sAPI()
	if err != nil {
		m.logger.WithError(err).Error("get data from k8s API")
		return err
	}
	fileHashesDto, err := m.repository.GetHashData(
		m.monitoringDirectory,
		m.algorithm,
		k8sData.DeploymentData,
	)
	if err != nil {
		return fmt.Errorf("failed get hash data: %w", err)
	}

	referenceHashes := make(map[string]*models.HashDataFromDB, len(fileHashesDto))

	for _, fh := range fileHashesDto {
		referenceHashes[fh.FullFilePath] = fh
	}

	for _, fh := range fileHashes {
		if fhdto, ok := referenceHashes[fh.Path]; ok {
			if fhdto.Hash != fh.Hash {
				m.integrityCheckFailed(
					ErrIntegrityFileMismatch,
					fh.Path,
					k8sData.KuberData,
					k8sData.DeploymentData,
				)
				return ErrIntegrityFileMismatch
			}
			delete(referenceHashes, fh.Path)
		} else {
			m.integrityCheckFailed(ErrIntegrityNewFileFoud, fh.Path, k8sData.KuberData, k8sData.DeploymentData)
			return ErrIntegrityNewFileFoud
		}
	}

	if len(referenceHashes) > 0 {
		for path := range referenceHashes {
			m.integrityCheckFailed(ErrIntegrityFileDeleted, path, k8sData.KuberData, k8sData.DeploymentData)
			return ErrIntegrityFileDeleted
		}
	}

	m.logger.Debug("end check integrity")
	return err
}

func (m *IntegrityMonitor) integrityCheckFailed(
	err error,
	path string,
	kubeData *models.KuberData,
	deploymentData *models.DeploymentData,
) {
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
	services.NewKuberService(m.logger).RolloutDeployment(kubeData)
}

func GetProcessPath(procName string, path string) (string, error) {
	pid, err := process.GetPID(procName)
	if err != nil {
		return "", fmt.Errorf("failed build process path: %w", err)
	}
	return fmt.Sprintf("/proc/%d/root/%s", pid, path), nil
}

func saveHashes(
	ctx context.Context,
	hashC <-chan filehash.FileHash,
	dd *models.DeploymentData,
	errC chan<- error,
) <-chan int {
	doneC := make(chan int)

	go func() {
		defer close(doneC)

		const defaultHashCnt = 100
		hashData := make([]*api.HashData, 0, defaultHashCnt)
		alg := viper.GetString("algorithm")
		countHashes := 0

		for v := range hashC {
			select {
			case <-ctx.Done():
				return
			default:
			}

			hashData = append(hashData, FileHashDtoDB(alg, &v))
			countHashes++
		}

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		query, args := data.NewHashFileData().PrepareBatchQuery(hashData, dd)
		err := repositories.ExecQueryTx(ctx, query, args...)
		if err != nil {
			errC <- err
		}
		doneC <- countHashes
	}()

	return doneC
}

func FileHashDtoDB(algName string, fh *filehash.FileHash) *api.HashData {
	return &api.HashData{
		Hash:         fh.Hash,
		FileName:     path.Base(fh.Path),
		FullFilePath: fh.Path,
		Algorithm:    algName,
	}
}
