package integritymonitor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/data"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/process"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/walker"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/worker"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/minio"
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

func CheckIntegrity(ctx context.Context, log *logrus.Logger, processName string, monitoringDirectories []string,
	deploymentData *k8s.DeploymentData, kubeClient *k8s.KubeClient) error {
	log.Debug("begin check integrity")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errC := make(chan error)
	defer close(errC)

	var err error
	paths := make([]string, len(monitoringDirectories))
	for i, p := range monitoringDirectories {
		paths[i], err = GetProcessPath(processName, p)
		if err != nil {
			log.WithError(err).Error("failed build process path")
			return err
		}

	}

	comparedHashesChan := compareHashes(ctx, log, worker.WorkersPool(
		viper.GetInt("count-workers"),
		walker.ChanWalkDir(ctx, paths, log),
		worker.NewWorker(ctx, viper.GetString("algorithm"), log),
	), processName, viper.GetString("algorithm"), deploymentData, errC)

	log.Trace("calculate & save hashes...")
	select {
	case <-ctx.Done():
		log.Error(ctx.Err())
		return ctx.Err()
	case countHashes := <-comparedHashesChan:
		log.WithField("countHashes", countHashes).Info("hashes compared successfully")
		return nil
	case err := <-errC:
		integrityCheckFailed(log, err, deploymentData, kubeClient)
		return err
	}
}

func compareHashes(
	ctx context.Context,
	log *logrus.Logger,
	hashC <-chan worker.FileHash,
	procName string,
	algName string,
	deploymentData *k8s.DeploymentData,
	errC chan<- error) <-chan int {

	doneC := make(chan int)
	go func() {
		defer close(doneC)

		procDirs, err := GetProcessPath(procName, "")
		if err != nil {
			errC <- fmt.Errorf("failed get process path: %w", err)
			return
		}

		ms := minio.Instance()
		csFile := process.CheckSumFile(procName, algName)
		log.Infof("getting check sums file %s", csFile)
		hashData, err := ms.Load(ctx, viper.GetString("minio-bucket"), csFile)
		if err != nil {
			errC <- fmt.Errorf("cannot read hash data: %w", err)
			return
		}

		expectedHashes, err := data.NewFileStorage(bytes.NewReader(hashData)).Get()
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

			strippedPaths := strings.TrimPrefix(v.Path, procDirs)
			if h, ok := expectedHashesMap[strippedPaths]; ok {
				if h != v.Hash {
					errC <- &IntegrityError{Type: ErrTypeFileMismatch, Path: strippedPaths, Hash: v.Hash}
					return
				}
				delete(expectedHashesMap, strippedPaths)
			} else {
				errC <- &IntegrityError{Type: ErrTypeNewFile, Path: strippedPaths, Hash: v.Hash}
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

	e := alerts.Send(alerts.New(fmt.Sprintf("Restart pod %v", deploymentData.NamePod),
		err.Error(),
		mPath,
	))
	if e != nil {
		log.WithError(e).Error("Failed send alert")
	}

	kubeClient.RestartPod()
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
