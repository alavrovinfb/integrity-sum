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

const (
	IntegrityMessageNewFileFound = "new file found"
	IntegrityMessageFileDeleted  = "file deleted"
	IntegrityMessageFileMismatch = "file content mismatch"
)

func GetProcessPath(procName string, path string) (string, error) {
	pid, err := process.GetPID(procName)
	if err != nil {
		return "", fmt.Errorf("failed build process path: %w", err)
	}
	return fmt.Sprintf("/proc/%d/root/%s", pid, path), nil
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

func CheckIntegrity(ctx context.Context, monitoringDirectory string, log *logrus.Logger, alertSender alerts.Sender, repo ports.IAppRepository) error {
	log.Debug("begin check integrity")

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

		case countHashes := <-compareHashes(
			ctx,
			worker.WorkersPool(
				viper.GetInt("count-workers"),
				walker.ChanWalkDir(ctx, monitoringDirectory, log),
				worker.NewWorker(ctx, viper.GetString("algorithm"), log),
			),
			repo,
			monitoringDirectory,
			viper.GetString("algorithm"),
			dataK8s.DeploymentData,
			errC,
		):
			log.WithField("countHashes", countHashes).Info("hashes compared successfully")
			log.Debug("end check integrity")
			return nil

		case err := <-errC:
			integrityCheckFailed(err, log, alertSender, dataK8s.KuberData, dataK8s.DeploymentData)
			return err
		}
	}
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

			hashData = append(hashData, fileHashDtoDB(alg, &v))
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

func compareHashes(
	ctx context.Context,
	hashC <-chan filehash.FileHash,
	repository ports.IAppRepository,
	directory string,
	algName string,
	deploymentData *models.DeploymentData,
	errC chan<- error,
) <-chan int {
	doneC := make(chan int)
	go func() {
		defer close(doneC)

		fileHashesDto, err := repository.GetHashData(
			directory,
			algName,
			deploymentData,
		)
		if err != nil {
			errC <- fmt.Errorf("failed get hash data: %w", err)
			return
		}

		//convert hashes to map
		hashesMap := make(map[string]string)
		for _, h := range fileHashesDto {
			hashesMap[h.FullFilePath] = h.Hash
		}

		for v := range hashC {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if h, ok := hashesMap[v.Path]; ok {
				if h != v.Hash {
					errC <- &IntegrityError{Type: ErrTypeFileMismatch, Path: v.Path, Hash: v.Hash}
					return
				}
				delete(hashesMap, v.Path)
			} else {
				errC <- &IntegrityError{Type: ErrTypeNewFile, Path: v.Path, Hash: v.Hash}
				return
			}
		}
		for p, h := range hashesMap {
			errC <- &IntegrityError{Type: ErrTypeFileDeleted, Path: p, Hash: h}
			return
		}
		doneC <- len(fileHashesDto)
	}()
	return doneC
}

func integrityCheckFailed(
	err error,
	log *logrus.Logger,
	alertSender alerts.Sender,
	kubeData *models.KuberData,
	deploymentData *models.DeploymentData,
) {
	var msg string
	var path string
	var integrityError *IntegrityError
	if errors.As(err, &integrityError) {
		path = integrityError.Path
		switch integrityError.Type {
		case ErrTypeFileMismatch:
			msg = IntegrityMessageFileMismatch
		case ErrTypeNewFile:
			msg = IntegrityMessageNewFileFound
		case ErrTypeFileDeleted:
			msg = IntegrityMessageFileDeleted
		}
		log.WithError(err).WithField("path", integrityError.Path).Error("check integrity failed")
	} else {
		msg = err.Error()
		log.WithError(err).Error("check integrity failed")
	}

	if alertSender != nil {
		err := alertSender.Send(alerts.Alert{
			Time:    time.Now(),
			Message: fmt.Sprintf("Restart deployment %v", deploymentData.NameDeployment),
			Reason:  msg,
			Path:    path,
		})
		if err != nil {
			log.WithError(err).Error("Failed send alert")
		}
	}
	services.NewKuberService(log).RolloutDeployment(kubeData)
}

func fileHashDtoDB(algName string, fh *filehash.FileHash) *api.HashData {
	return &api.HashData{
		Hash:         fh.Hash,
		FileName:     path.Base(fh.Path),
		FullFilePath: fh.Path,
		Algorithm:    algName,
	}
}
