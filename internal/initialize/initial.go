package initialize

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/integrity-sum/internal/core/services"
	"github.com/integrity-sum/internal/repositories"
	"github.com/integrity-sum/pkg/notification"
	"github.com/integrity-sum/pkg/notification/splunkclient"
)

// config defaults
const (
	procDir                = "/proc"
	durationTime           = 30 * time.Second
	algorithm              = "SHA256"
	procToMonitor          = "sh" // just a placeholder must be set
	pathToMonitor          = "/"
)

func init() {
	fsSum := pflag.NewFlagSet("sum", pflag.ContinueOnError)
	fsSum.String("proc-dir", procDir, "path to /proc")
	fsSum.Duration("duration-time", durationTime, "specific interval of time repeatedly for ticker")
	fsSum.Int("count-workers", runtime.NumCPU(), "number of running workers in the workerpool")
	fsSum.String("algorithm", algorithm, "hashing algorithm for hashing data")
	fsSum.String("process", procToMonitor, "the name of the process to be monitored by the hasher")
	fsSum.String("monitoring-path", pathToMonitor, "the service path to be monitored by the hasher")
	pflag.CommandLine.AddFlagSet(fsSum)
	if err := viper.BindPFlags(fsSum); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Initialize(ctx context.Context, logger *logrus.Logger, sig chan os.Signal) {
	// Initialize repository

	db, err := repositories.ConnectionToDB(logger)
	if err != nil {
		logger.Fatalf("can't connect to database: %s", err)
	}

	repository := repositories.NewAppRepository(logger, db)

	notifier := splunkclient.New(logger, "http://splunk:8088/services/collector/event", "72fbe9ab-2b51-4784-bf07-c2fe96489be1", true)
	err = notifier.Send(notification.Message{Time: time.Now(), Message: "test message"})
	if err != nil {
		logger.WithError(err).Debug("Error Send notification")
	}

	// Initialize service
	algorithm := viper.GetString("algorithm")

	service := services.NewAppService(repository, algorithm, logger)

	// Initialize kubernetesAPI
	dataFromK8sAPI, err := service.GetDataFromK8sAPI()
	if err != nil {
		logger.Fatalf("can't get data from K8sAPI: %s", err)
	}

	//Getting pid
	procName := viper.GetString("process")
	pid, err := service.GetPID(procName)
	if err != nil {
		logger.Fatalf("err while getting pid %s", err)
	}
	if pid == 0 {
		logger.Fatalf("proc with name %s not exist", procName)
	}

	//Getting the path to the monitoring directory
	dirPath := fmt.Sprintf("/proc/%d/root/%s", pid, viper.GetString("monitoring-path"))

	ticker := time.NewTicker(viper.GetDuration("duration-time"))

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context, ticker *time.Ticker) {
		defer wg.Done()
		for {
			if !service.IsExistDeploymentNameInDB(dataFromK8sAPI.KuberData.TargetName) {
				logger.Info("Deployment name does not exist in database, save data")
				err = service.Start(ctx, dirPath, sig, dataFromK8sAPI.DeploymentData)
				if err != nil {
					logger.Fatalf("Error when starting to get and save hash data %s", err)
				}
			} else {
				logger.Info("Deployment name exists in database, checking data")
				for range ticker.C {
					err = service.Check(ctx, dirPath, sig, dataFromK8sAPI.DeploymentData, dataFromK8sAPI.KuberData)
					if err != nil {
						logger.Fatalf("Error when starting to check hash data %s", err)
					}
					logger.Info("Check completed")
				}
			}
		}
	}(ctx, ticker)
	wg.Wait()
	ticker.Stop()
}
