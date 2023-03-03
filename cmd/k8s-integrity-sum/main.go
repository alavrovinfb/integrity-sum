package main

import (
	"context"
	"flag"
	"fmt"

	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/configs"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/services"
	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/ffi/bee2" // bee2 registration
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
	log := logger.Init(viper.GetString("verbose"))

	// Install migration
	DBMigration(log)

	monitor, err := initMonitor(log)
	if err != nil {
		log.WithError(err).Fatal("failed initialize integrity monitor")
	}

	// Run Application with graceful shutdown context
	graceful.Execute(context.Background(), log, func(ctx context.Context) {
		err := monitor.Run(ctx)
		if err == context.Canceled {
			log.Info("execution cancelled")
			return
		}
		if err != nil {
			log.WithError(err).Error("monitor execution aborted")
			return
		}
	})
}

func initMonitor(log *logrus.Logger) (*integritymonitor.IntegrityMonitor, error) {
	// Initialize database
	db, err := repositories.ConnectionToDB(log)
	if err != nil {
		return nil, fmt.Errorf("failed connect to database: %w", err)
	}
	repository := repositories.NewAppRepository(log, db)

	// Create alert sender
	splunkUrl := viper.GetString("splunk-url")
	splunkToken := viper.GetString("splunk-token")
	splunkInsecureSkipVerify := viper.GetBool("splunk-insecure-skip-verify")
	var alertsSender alerts.Sender
	if len(splunkUrl) > 0 && len(splunkToken) > 0 {
		alertsSender = splunk.New(log, splunkUrl, splunkToken, splunkInsecureSkipVerify)
	}

	// Kube client
	kubeClient := services.NewKuberService(log)
	// kube client connection must be placed here with error handling

	// Initialize service
	algorithm := viper.GetString("algorithm")
	countWorkers := viper.GetInt("count-workers")
	fileHasher := filehash.NewFileSystemHasher(log, algorithm, countWorkers)

	monitorDelay := viper.GetDuration("duration-time")
	monitorProc := viper.GetString("process")
	monitorPath := viper.GetString("monitoring-path")
	return integritymonitor.New(log, fileHasher, repository, kubeClient, alertsSender, monitorDelay, monitorProc, monitorPath, algorithm)
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
