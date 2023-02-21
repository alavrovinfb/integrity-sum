package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/initialize"
	logConfig "github.com/ScienceSoft-Inc/integrity-sum/pkg/logger"
)

func main() {
	// Install config
	initConfig()

	// Install logger
	logger := logConfig.InitLogger(viper.GetInt("verbose"))

	// Install migration
	DBMigration(logger)

	// Handling shutdown signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		signal.Stop(sig)
		cancel()
	}()

	// Initialize program
	initialize.Initialize(ctx, logger, sig)
}

func initConfig() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.AutomaticEnv()
}
