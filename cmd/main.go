package main

import (
	"context"
	logConfig "github.com/integrity-sum/pkg/logger"
	"log"
	"os"
	"os/signal"

	"github.com/integrity-sum/internal/initialize"
	"github.com/joho/godotenv"
)

func main() {
	// Load values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
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
