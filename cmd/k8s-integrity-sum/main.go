package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/configs"
	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/ffi/bee2"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/logger"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehash"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/integritymonitor"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/graceful"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts/splunk"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/common"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/health"
)

func main() {
	// Install config
	initConfig()

	// Install logger
	log := logger.Init(viper.GetString("verbose"))

	// Set app health status to healthy
	h := health.New(fmt.Sprintf("/tmp/%s", common.AppId))
	err := h.Set()
	if err != nil {
		log.Fatalf("cannot create health file")
	}
	defer h.Reset()

	// Install migration
	DBMigration(log)

	// DB connect
	if _, err := repositories.Open(log); err != nil {
		log.Fatalf("failed connect to database: %w", err)
	}

	monitor, err := initMonitor(log)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize integrity monitor")
	}

	// Run Application with graceful shutdown context
	graceful.Execute(context.Background(), log, func(ctx context.Context) {
		if err = setupIntegrity(ctx, log); err != nil {
			log.WithError(err).Fatal("failed to setup integrity")
			return
		}

		// TODO: make it independent
		err := monitor.Run(ctx, viper.GetDuration("duration-time"), viper.GetString("algorithm"))
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
	// Create alert sender
	splunkUrl := viper.GetString("splunk-url")
	splunkToken := viper.GetString("splunk-token")
	splunkInsecureSkipVerify := viper.GetBool("splunk-insecure-skip-verify")
	var alertsSender alerts.Sender
	if len(splunkUrl) > 0 && len(splunkToken) > 0 {
		alertsSender = splunk.New(log, splunkUrl, splunkToken, splunkInsecureSkipVerify)
	}

	// Initialize service
	algorithm := viper.GetString("algorithm")
	countWorkers := viper.GetInt("count-workers")
	fileHasher := filehash.NewFileSystemHasher(log, algorithm, countWorkers) // TODO: remove

	monitorProc := viper.GetString("process")
	monitorPath := viper.GetString("monitoring-path")
	return integritymonitor.New(log, fileHasher, alertsSender, monitorProc, monitorPath)
}

func setupIntegrity(ctx context.Context, log *logrus.Logger) error {
	processPath, err := integritymonitor.GetProcessPath(
		viper.GetString("process"),
		viper.GetString("monitoring-path"),
	)
	if err != nil {
		return err
	}
	return integritymonitor.SetupIntegrity(ctx, processPath, log)
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
