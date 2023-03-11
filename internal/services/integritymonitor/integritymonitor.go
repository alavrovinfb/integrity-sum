package integritymonitor

import (
	"context"
	"errors"
	"fmt"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/data"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehash"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/process"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/walker"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/worker"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
)

var ErrIntegrityNewFileFoud = errors.New("new file found")
var ErrIntegrityFileDeleted = errors.New("file deleted")
var ErrIntegrityFileMismatch = errors.New("file content mismatch")

type IntegrityMonitor struct {
	logger              *logrus.Logger
	fshasher            *filehash.FileSystemHasher
	alertSender         alerts.Sender
	monitoringDirectory string
}

func New(logger *logrus.Logger,
	fshasher *filehash.FileSystemHasher,
	alertSender alerts.Sender,
	monitorProcess string,
	monitorProcessPath string,
) (*IntegrityMonitor, error) {
	processPath, err := GetProcessPath(monitorProcess, monitorProcessPath)
	if err != nil {
		return nil, err
	}

	return &IntegrityMonitor{
		logger:              logger,
		fshasher:            fshasher,
		alertSender:         alertSender,
		monitoringDirectory: processPath,
	}, nil
}

func (m *IntegrityMonitor) Run(ctx context.Context, interval time.Duration, algName string) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ticker.Stop()
			err := m.checkIntegrity(ctx, algName)
			if err != nil {
				m.logger.WithError(err).Error("failed check integrity")
			}
			ticker.Reset(interval)
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

	dataK8s, err := k8s.NewKubeService(log).GetDataFromK8sAPI()
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
			log,
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

func (m *IntegrityMonitor) checkIntegrity(ctx context.Context, algName string) error {
	m.logger.Debug("begin check integrity")
	fileHashes, err := m.fshasher.CalculateAll(ctx, m.monitoringDirectory)
	if err != nil {
		m.logger.WithError(err).Error("failed calculate file hashes")
		return err
	}

	k8sData, err := k8s.NewKubeService(m.logger).GetDataFromK8sAPI()
	if err != nil {
		m.logger.WithError(err).Error("get data from k8s API")
		return err
	}

	err = data.NewReleaseStorage(
		data.DB().SQL(),
		algName,
		m.logger,
	).Update(k8sData.DeploymentData.NameDeployment)
	if err != nil {
		return fmt.Errorf("failed update releases data: %w", err)
	}

	fileHashesDto, err := data.NewHashStorage(
		data.DB().SQL(),
		algName,
		m.logger,
	).Get(
		m.monitoringDirectory,
		k8sData.DeploymentData,
	)
	if err != nil {
		return fmt.Errorf("failed get hash data: %w", err)
	}

	referenceHashes := make(map[string]*data.HashData, len(fileHashesDto))

	for _, fh := range fileHashesDto {
		referenceHashes[fh.FullFileName] = fh
	}

	for _, fh := range fileHashes {
		if fhdto, ok := referenceHashes[fh.Path]; ok {
			if fhdto.Hash != fh.Hash {
				m.integrityCheckFailed(
					ErrIntegrityFileMismatch,
					fh.Path,
					k8sData.KubeData,
					k8sData.DeploymentData,
				)
				return ErrIntegrityFileMismatch
			}
			delete(referenceHashes, fh.Path)
		} else {
			m.integrityCheckFailed(ErrIntegrityNewFileFoud, fh.Path, k8sData.KubeData, k8sData.DeploymentData)
			return ErrIntegrityNewFileFoud
		}
	}

	if len(referenceHashes) > 0 {
		for path := range referenceHashes {
			m.integrityCheckFailed(
				ErrIntegrityFileDeleted,
				path,
				k8sData.KubeData,
				k8sData.DeploymentData,
			)
			return ErrIntegrityFileDeleted
		}
	}

	m.logger.Debug("end check integrity")
	return err
}

func (m *IntegrityMonitor) integrityCheckFailed(
	err error,
	path string,
	kubeData *k8s.KubeData,
	deploymentData *k8s.DeploymentData,
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
	k8s.NewKubeService(m.logger).RolloutDeployment(kubeData)
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
	dd *k8s.DeploymentData,
	errC chan<- error,
	logger *logrus.Logger,
) <-chan int {
	doneC := make(chan int)

	go func() {
		defer close(doneC)

		const defaultHashCnt = 100
		hashData := make([]*data.HashData, 0, defaultHashCnt)
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

		//ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		//defer cancel()

		//query, args := data.NewHashFileData().PrepareBatchQuery(hashData, dd)
		err := data.NewReleaseStorage(data.DB().SQL(), alg, logger).Create(dd)
		//err := repositories.ExecQueryTx(ctx, query, args...)
		if err != nil {
			errC <- err
		}
		err = data.NewHashStorage(data.DB().SQL(), alg, logger).Create(hashData, dd)
		//err := repositories.ExecQueryTx(ctx, query, args...)
		if err != nil {
			errC <- err
		}
		doneC <- countHashes
	}()

	return doneC
}

func FileHashDtoDB(algName string, fh *filehash.FileHash) *data.HashData {
	return &data.HashData{
		Hash:         fh.Hash,
		FullFileName: fh.Path,
		Algorithm:    algName,
	}
}
