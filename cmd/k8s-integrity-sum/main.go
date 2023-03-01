package main

import (
	"context"
	"flag"

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

	// Run Application with graceful shutdown context
	graceful.Execute(context.Background(), logger, func(ctx context.Context) {

		// initialize.Initialize(ctx, logger)

		monitor := initMonitor(ctx, logger)
		err := monitor.Run(ctx)
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			logger.WithError(err).Error("monitor execution completed with error")
		}
	})
}

func initMonitor(ctx context.Context, logger *logrus.Logger) *integritymonitor.IntegrityMonitor {
	// Initialize database
	db, err := repositories.ConnectionToDB(logger)
	if err != nil {
		logger.Fatalf("can't connect to database: %s", err)
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

	// Initialize service
	algorithm := viper.GetString("algorithm")

	countWorkers := viper.GetInt("count-workers")
	monitorDelay := viper.GetDuration("duration-time")
	monitorProc := viper.GetString("process")
	monitorPath := viper.GetString("monitoring-path")

	fileHasher := filehash.NewFileSystemHasher(logger, algorithm, countWorkers)

	return integritymonitor.New(logger, fileHasher, repository, kubeClient, alertsSender, monitorDelay, monitorProc, monitorPath, algorithm)
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
