package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/integrity-sum/internal/core/services"
	"github.com/integrity-sum/internal/repositories"
	"github.com/integrity-sum/pkg/api"
	logConfig "github.com/integrity-sum/pkg/logger"
	"os"
	"os/signal"
)

var dirPath string
var algorithm string
var doHelp bool

// Initializes the binding of the flag to a variable that must run before the main() function
func init() {
	flag.StringVar(&dirPath, "d", "", "a specific file or directory")
	flag.StringVar(&algorithm, "a", "SHA256", "algorithm MD5, SHA1, SHA224, SHA256, SHA384, SHA512, default: SHA256")
	flag.BoolVar(&doHelp, "h", false, "help")
}

func main() {
	flag.Parse()

	// Initialize config for logger
	logger, err := logConfig.LoadConfig()
	if err != nil {
		logger.Fatal("Error during loading from config file", err)
	}

	// Install context and signal
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	defer func() {
		signal.Stop(sig)
		cancel()
	}()

	switch {
	case doHelp:
		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, "Custom help %s:\nYou can use the following flag:\n", os.Args[0])

			flag.VisitAll(func(f *flag.Flag) {
				fmt.Fprintf(os.Stderr, "  flag -%v \n       %v\n", f.Name, f.Usage)
			})
		}
		flag.Usage()
	case len(dirPath) > 0:
		// Initialize repository
		repository := repositories.NewAppRepository(logger)

		// Initialize service
		service := services.NewAppService(repository, algorithm, logger)
		if err != nil {
			logger.Fatalf("can't init service: %s", err)
		}

		jobs := make(chan string)
		results := make(chan *api.HashData)

		go service.WorkerPool(jobs, results)
		go api.SearchFilePath(dirPath, jobs, logger)
		for {
			select {
			case hashData, ok := <-results:
				if !ok {
					return
				}
				fmt.Printf("%s %s\n", hashData.Hash, hashData.FileName)
			case <-sig:
				fmt.Println("exit program")
				return
			case <-ctx.Done():
				fmt.Println("program termination after receiving a signal")
				return
			}
		}
	default:
		logger.Println("use the -h flag on the command line to see all the flags in this app")
	}
}
