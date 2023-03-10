package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/configs"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/services"
	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/ffi/bee2"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/integritymonitor"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/logger"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/repositories"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/graceful"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts/splunk"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/common"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/health"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

	// 	// Create alert sender
	splunkUrl := viper.GetString("splunk-url")
	splunkToken := viper.GetString("splunk-token")
	splunkInsecureSkipVerify := viper.GetBool("splunk-insecure-skip-verify")
	var alertsSender alerts.Sender
	if len(splunkUrl) > 0 && len(splunkToken) > 0 {
		alertsSender = splunk.New(log, splunkUrl, splunkToken, splunkInsecureSkipVerify)
	}

	kubeClient := services.NewKubeService(log)
	_, err = kubeClient.Connect()
	if err != nil {
		log.Fatalf("failed connect to kubernetes: %w", err)
	}
	kubeData, err := kubeClient.GetKubeData()
	if err != nil {
		log.Fatalf("failed get kube data: %w", err)
	}
	deploymentData, err := kubeClient.GetDataFromDeployment(kubeData)
	if err != nil {
		log.Fatalf("failed get deployment data: %w", err)
	}

	// Run Application with graceful shutdown context
	graceful.Execute(context.Background(), log, func(ctx context.Context) {
		if err = setupIntegrity(ctx, log, deploymentData); err != nil {
			log.WithError(err).Fatal("failed to setup integrity")
			return
		}

		err := runCheckIntegrity(ctx, log, alertsSender, deploymentData, kubeClient)
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

func setupIntegrity(ctx context.Context, log *logrus.Logger, deploymentData *models.DeploymentData) error {
	processPath, err := integritymonitor.GetProcessPath(
		viper.GetString("process"),
		viper.GetString("monitoring-path"),
	)
	if err != nil {
		return err
	}
	return integritymonitor.SetupIntegrity(ctx, processPath, log, deploymentData)
}

func runCheckIntegrity(ctx context.Context, log *logrus.Logger, alertSender alerts.Sender, deploymentData *models.DeploymentData, kubeClient *services.KubeClient) error {

	processPath, err := integritymonitor.GetProcessPath(
		viper.GetString("process"),
		viper.GetString("monitoring-path"),
	)
	if err != nil {
		return err
	}

	t := time.NewTicker(viper.GetDuration("duration-time"))
	for range t.C {
		err := integritymonitor.CheckIntegrity(ctx, log, processPath, alertSender, deploymentData, kubeClient)
		if err != nil {
			return err
		}
	}
	return nil
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
