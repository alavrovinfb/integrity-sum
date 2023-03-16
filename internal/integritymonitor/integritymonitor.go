package integritymonitor

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/services"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories/data"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/process"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/walker"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/worker"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

func SetupIntegrity(ctx context.Context, monitoringDirectory string, log *logrus.Logger, deploymentData *models.DeploymentData) error {
	log.Debug("begin setup integrity")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errC := make(chan error)
	defer close(errC)

	log.Trace("calculate & save hashes...")
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
		deploymentData,
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

func CheckIntegrity(ctx context.Context,
	log *logrus.Logger,
	monitoringDirectory string,
	kubeData *models.KubeData,
	deploymentData *models.DeploymentData,
	kubeClient *services.KubeClient) error {
	log.Debug("begin check integrity")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errC := make(chan error)
	defer close(errC)

	comparedHashesChan := compareHashes(
		ctx,
		log,
		worker.WorkersPool(
			viper.GetInt("count-workers"),
			walker.ChanWalkDir(ctx, monitoringDirectory, log),
			worker.NewWorker(ctx, viper.GetString("algorithm"), log),
		),
		monitoringDirectory,
		viper.GetString("algorithm"),
		deploymentData,
		errC,
	)

	log.Trace("calculate & save hashes...")
	select {
	case <-ctx.Done():
		log.Error(ctx.Err())
		return ctx.Err()
	case countHashes := <-comparedHashesChan:
		log.WithField("countHashes", countHashes).Info("hashes compared successfully")
		return nil
	case err := <-errC:
		integrityCheckFailed(log, err, kubeData, deploymentData, kubeClient)
		return err
	}
}

func saveHashes(
	ctx context.Context,
	hashC <-chan worker.FileHash,
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

			hashData = append(hashData, fileHashToDtoDB(alg, &v))
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
	log *logrus.Logger,
	hashC <-chan worker.FileHash,
	directory string,
	algName string,
	deploymentData *models.DeploymentData,
	errC chan<- error,
) <-chan int {
	doneC := make(chan int)
	go func() {
		defer close(doneC)

		repository := repositories.NewAppRepository(log, repositories.DB().SQL())

		expectedHashes, err := repository.GetHashData(
			directory,
			algName,
			deploymentData,
		)
		if err != nil {
			errC <- fmt.Errorf("failed get hash data: %w", err)
			return
		}

		//convert hashes to map
		expectedHashesMap := make(map[string]string)
		for _, h := range expectedHashes {
			expectedHashesMap[h.FullFilePath] = h.Hash
		}

		for v := range hashC {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if h, ok := expectedHashesMap[v.Path]; ok {
				if h != v.Hash {
					errC <- &IntegrityError{Type: ErrTypeFileMismatch, Path: v.Path, Hash: v.Hash}
					return
				}
				delete(expectedHashesMap, v.Path)
			} else {
				errC <- &IntegrityError{Type: ErrTypeNewFile, Path: v.Path, Hash: v.Hash}
				return
			}
		}
		for p, h := range expectedHashesMap {
			errC <- &IntegrityError{Type: ErrTypeFileDeleted, Path: p, Hash: h}
			return
		}
		doneC <- len(expectedHashes)
	}()
	return doneC
}

func integrityCheckFailed(
	log *logrus.Logger,
	err error,
	kubeData *models.KubeData,
	deploymentData *models.DeploymentData,
	kubeClient *services.KubeClient,
) {
	l := log.WithError(err)
	var path string
	var integrityError *IntegrityError
	if errors.As(err, &integrityError) {
		path = integrityError.Path
		l = l.WithField("path", integrityError.Path)
	}

	l.Error("check integrity failed")

	e := alerts.Send(alerts.Alert{
		Time:    time.Now(),
		Message: fmt.Sprintf("Restart pod %v", deploymentData.NamePod),
		Reason:  err.Error(),
		Path:    path,
	})
	if e != nil {
		log.WithError(e).Error("Failed send alert")
	}

	kubeClient.RestartPod(kubeData)
}

func fileHashToDtoDB(algName string, fh *worker.FileHash) *api.HashData {
	return &api.HashData{
		Hash:         fh.Hash,
		FileName:     path.Base(fh.Path),
		FullFilePath: fh.Path,
		Algorithm:    algName,
	}
}

func ParseMonitoringOpts(opts string) (map[string][]string, error) {
	if opts == "" {
		return nil, fmt.Errorf("--%s %s", "monitoring-options", "is empty")
	}
	unOpts, err := strconv.Unquote(opts)
	if err != nil {
		unOpts = opts
	}

	processes := strings.Split(unOpts, " ")
	if len(processes) < 1 {
		return nil, fmt.Errorf("--%s %s", "monitoring-options", "is empty")
	}
	optsMap := make(map[string][]string)
	for _, p := range processes {
		procPaths := strings.Split(p, "=")
		if len(procPaths) < 2 {
			return nil, fmt.Errorf("%s", "application and monitoring paths should be represented as key=value pair")
		}

		if procPaths[1] == "" {
			return nil, fmt.Errorf("%s", "monitoring path is required")
		}
		paths := strings.Split(strings.Trim(procPaths[1], ","), ",")
		for i, v := range paths {
			paths[i] = strings.TrimSpace(v)
		}
		optsMap[procPaths[0]] = paths
	}
	return optsMap, nil
}
