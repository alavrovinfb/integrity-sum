package integritymonitor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/data"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/process"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/walker"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/worker"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s"
)

const (
	IntegrityMessageNewFileFound = "new file found"
	IntegrityMessageFileDeleted  = "file deleted"
	IntegrityMessageFileMismatch = "file content mismatch"
	IntegrityMessageUnknownErr   = "unknown integrity error"
)

func GetProcessPath(procName string, path string) (string, error) {
	pid, err := process.GetPID(procName)
	if err != nil {
		return "", fmt.Errorf("failed build process path: %w", err)
	}
	return fmt.Sprintf("/proc/%d/root/%s", pid, path), nil
}

func SetupIntegrity(ctx context.Context, monitoringDirectories []string, log *logrus.Logger, deploymentData *k8s.DeploymentData) error {
	log.Debug("begin setup integrity")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errC := make(chan error)
	defer close(errC)

	saveHahesChan := saveHashes(
		ctx,
		log,
		worker.WorkersPool(
			viper.GetInt("count-workers"),
			walker.ChanWalkDir(ctx, monitoringDirectories, log),
			worker.NewWorker(ctx, viper.GetString("algorithm"), log),
		),
		deploymentData,
		errC,
	)

	log.Trace("calculate & save hashes...")
	select {
	case <-ctx.Done():
		log.Error(ctx.Err())
		return ctx.Err()

	case countHashes := <-saveHahesChan:
		log.WithField("countHashes", countHashes).WithField("monitoringDirectories", monitoringDirectories).Info("hashes stored successfully")
		log.Debug("end setup integrity")
		return nil

	case err := <-errC:
		log.WithError(err).Error("setup integrity failed")
		return err
	}
}

func CheckIntegrity(ctx context.Context, log *logrus.Logger, monitoringDirectories []string, kubeData *k8s.KubeData, deploymentData *k8s.DeploymentData, kubeClient *k8s.KubeClient) error {
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
			walker.ChanWalkDir(ctx, monitoringDirectories, log),
			worker.NewWorker(ctx, viper.GetString("algorithm"), log),
		),
		monitoringDirectories[0],
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
	log *logrus.Logger,
	hashC <-chan worker.FileHash,
	dd *k8s.DeploymentData,
	errC chan<- error,
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

			hashData = append(hashData, fileHashToDtoDB(v, alg, dd.NamePod, 0))
			countHashes++
		}

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		queryR, argsR := data.NewReleaseData(data.DB().SQL()).PrepareQuery(dd.NameDeployment, dd.Image, "deployment")
		queryH, argsH := data.NewHashData(data.DB().SQL()).PrepareQuery(hashData, dd.NameDeployment)
		err := data.WithTx(func(txn *sql.Tx) error {
			_, err := txn.ExecContext(ctx, queryR, argsR...)
			if err != nil {
				return err
			}
			_, err = txn.ExecContext(ctx, queryH, argsH...)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			errC <- err
			return
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
	deploymentData *k8s.DeploymentData,
	errC chan<- error,
) <-chan int {
	doneC := make(chan int)
	go func() {
		defer close(doneC)
		// TODO update with actual process
		procDir, err := GetProcessPath("nginx", "")
		if err != nil {
			errC <- fmt.Errorf("failed get hash data: %w", err)
			return
		}

		expectedHashes, err := data.NewHashData(data.DB().SQL()).Get(algName, procDir, deploymentData.NamePod)
		if err != nil {
			errC <- fmt.Errorf("failed get hash data: %w", err)
			return
		}

		//convert hashes to map
		expectedHashesMap := make(map[string]string)
		for _, h := range expectedHashes {
			expectedHashesMap[h.FullFileName] = h.Hash
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
	kubeData *k8s.KubeData,
	deploymentData *k8s.DeploymentData,
	kubeClient *k8s.KubeClient,
) {
	l := log.WithError(err)
	var mPath string
	var integrityError *IntegrityError
	if errors.As(err, &integrityError) {
		mPath = integrityError.Path
		l = l.WithField("path", integrityError.Path)
	}

	l.Error("check integrity failed")

	e := alerts.Send(alerts.New(fmt.Sprintf("Restart deployment %v", deploymentData.NameDeployment),
		err.Error(),
		mPath,
	))
	if e != nil {
		log.WithError(e).Error("Failed send alert")
	}

	kubeClient.RolloutDeployment(kubeData)
}

func fileHashToDtoDB(fh worker.FileHash, algName string, podName string, releaseId int) *data.HashData {
	return &data.HashData{
		Hash:         fh.Hash,
		FullFileName: fh.Path,
		Algorithm:    algName,
		PodName:      podName,
		ReleaseId:    releaseId,
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
