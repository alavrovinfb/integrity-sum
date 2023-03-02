package main

import (
	"context"
	"flag"
	"fmt"

	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/configs"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/services"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/logger"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehash"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/integritymonitor"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/graceful"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts/splunk"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// Install config
	initConfig()

	// Install logger
	logger := logger.Init(viper.GetString("verbose"))

	// Install migration
	DBMigration(logger)

	monitor, err := initMonitor(logger)
	if err != nil {
		logger.WithError(err).Fatal("failed initialize integrity monitor")
	}

	// Run Application with graceful shutdown context
	graceful.Execute(context.Background(), logger, func(ctx context.Context) {

		// Initialize function still can be used to run previous implementation
		// initialize.Initialize(ctx, logger)

		err := monitor.Run(ctx)
		if err == context.Canceled {
			logger.Info("execution cancelled")
			return
		}
		if err != nil {
			logger.WithError(err).Error("monitor execution aborted")
			return
		}
	})
}

func initMonitor(logger *logrus.Logger) (*integritymonitor.IntegrityMonitor, error) {
	// Initialize database
	db, err := repositories.ConnectionToDB(logger)
	if err != nil {
		return nil, fmt.Errorf("failed connect to database: %w", err)
	}
	repository := repositories.NewAppRepository(logger, db)

	// Create alert sender
	splunkUrl := viper.GetString("splunk-url")
	splunkToken := viper.GetString("splunk-token")
	splunkInsecureSkipVerify := viper.GetBool("splunk-insecure-skip-verify")
	var alertsSender alerts.Sender
	if len(splunkUrl) > 0 && len(splunkToken) > 0 {
		alertsSender = splunk.New(logger, splunkUrl, splunkToken, splunkInsecureSkipVerify)
	}

	// Kube client
	kubeClient := services.NewKuberService(logger)
	// kube client connection must be placed here with error handling

	// Initialize service
	algorithm := viper.GetString("algorithm")
	countWorkers := viper.GetInt("count-workers")
	fileHasher := filehash.NewFileSystemHasher(logger, algorithm, countWorkers)

	monitorDelay := viper.GetDuration("duration-time")
	monitorProc := viper.GetString("process")
	monitorPath := viper.GetString("monitoring-path")
	return integritymonitor.New(logger, fileHasher, repository, kubeClient, alertsSender, monitorDelay, monitorProc, monitorPath, algorithm), nil
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
