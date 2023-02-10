package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/integrity-sum/internal/initialize"
	logConfig "github.com/integrity-sum/pkg/logger"
)

func main() {
	initConfig()

	// Load values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	// Initialize config for logger
	logger, err := logConfig.LoadConfig()
	if err != nil {
		logger.Fatal("Error during loading from config file", err)
	}

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
	pflag.Parse()
	viper.AutomaticEnv()
	// fmt.Printf("dbconn: %v\n", configs.GetDBConnString())
}
