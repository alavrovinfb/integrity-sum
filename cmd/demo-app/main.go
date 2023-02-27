package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/logger"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehashservice"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/graceful"
)

// Initializes the binding of the flag to a variable that must run before the main() function
func init() {
	fsLog := pflag.NewFlagSet("log", pflag.ContinueOnError)
	fsLog.String("verbose", "info", "verbose level")

	fsSum := pflag.NewFlagSet("sum", pflag.ContinueOnError)
	fsSum.Int("count-workers", runtime.NumCPU(), "number of running workers in the workerpool")
	fsSum.String("algorithm", "SHA256", "algorithm MD5, SHA1, SHA224, SHA256, SHA384, SHA512, default: SHA256")
	fsSum.String("dirPath", "./", "name of configMap for hasher")
	fsSum.Bool("doHelp", false, "help")
	if err := viper.BindPFlags(fsSum); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	pflag.Parse()
	logger := logger.Init(viper.GetString("verbose"))
	graceful.Execute(context.Background(), logger, func(ctx context.Context) {
		run(ctx, logger)
	})
}

func run(ctx context.Context, logger *logrus.Logger) {
	switch {
	case viper.GetBool("doHelp"):
		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, "Custom help %s:\nYou can use the following flag:\n", os.Args[0])

			flag.VisitAll(func(f *flag.Flag) {
				fmt.Fprintf(os.Stderr, "  flag -%v \n       %v\n", f.Name, f.Usage)
			})
		}
		flag.Usage()
	case len(viper.GetString("dirPath")) > 0:
		hashservice := filehashservice.New(logger, viper.GetString("algorithm"), viper.GetString("dirPath"), viper.GetInt("count-workers"))
		resultChan, errChan := hashservice.CalculateInChan(ctx)

		for {
			select {
			case hashData, ok := <-resultChan:
				if ok {
					fmt.Printf("%s %s\n", hashData.Hash, hashData.Path)
				} else {
					resultChan = nil
				}
			case err := <-errChan:
				if err != nil {
					fmt.Printf("error: %v\n", err)
					return
				}
			case <-ctx.Done():
				fmt.Println("program termination after receiving a signal")
				return
			}
		}
	default:
		logger.Println("use the -h flag on the command line to see all the flags in this app")
	}
}
