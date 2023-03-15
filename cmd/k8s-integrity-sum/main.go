package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

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
		log.Fatalf("failed connect to database: %v", err)
	}

	optsMap, err := integritymonitor.ParseMonitoringOpts(viper.GetString("monitoring-options"))
	if err != nil {
		log.WithError(err).Fatal("cannot parse monitoring options")
	}
	monitors := make([]*integritymonitor.IntegrityMonitor, len(optsMap))
	i := 0
	for proc, paths := range optsMap {
		monitors[i], err = initMonitor(log, proc, paths)
		if err != nil {
			log.WithError(err).Fatal("failed to initialize integrity monitor")
		}
		i++
	}

	// Run Application with graceful shutdown context
	graceful.Execute(context.Background(), log, func(ctx context.Context) {
		if err = setupIntegrity(ctx, log, optsMap); err != nil {
			log.WithError(err).Fatal("failed to setup integrity")
			return
		}

		// TODO: make it independent
		g := errgroup.Group{}
		for _, monitor := range monitors {
			monitor := monitor
			g.Go(func() error {
				err := monitor.Run(ctx, viper.GetDuration("duration-time"), viper.GetString("algorithm"))
				if err == context.Canceled {
					log.Info("execution cancelled")
					return err
				}
				if err != nil {
					log.WithError(err).Error("monitor execution aborted")
					return err
				}
				return nil
			})
		}
		g.Wait()
	})
}

func initMonitor(log *logrus.Logger, procName string, procPaths []string) (*integritymonitor.IntegrityMonitor, error) {
	// TODO: separated: storage, data models; remove repository, remove repository dependency from the monitor.
	repository := repositories.NewAppRepository(log, repositories.DB().SQL())

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

	return integritymonitor.New(log, fileHasher, repository, alertsSender, procName, procPaths)
}

func setupIntegrity(ctx context.Context, log *logrus.Logger, optsMap map[string][]string) error {
	g := errgroup.Group{}
	for pName, pPaths := range optsMap {
		pName := pName
		pPaths := pPaths
		g.Go(func() error {
			for _, p := range pPaths {
				processPath, err := integritymonitor.GetProcessPath(pName, p)
				if err != nil {
					return err
				}
				if err := integritymonitor.SetupIntegrity(ctx, processPath, log); err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
