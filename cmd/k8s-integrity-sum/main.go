package main

import (
	"context"
	"flag"

	_ "github.com/ScienceSoft-Inc/integrity-sum/internal/configs"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/initialize"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/logger"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/graceful"
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
		initialize.Initialize(ctx, logger)
	})
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
