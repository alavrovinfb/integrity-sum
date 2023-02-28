package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func Execute(pCtx context.Context, logger *logrus.Logger, execute func(context.Context)) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(pCtx)
	go func() {
		if s, ok := <-sig; ok {
			logger.WithField("signal", s.String()).Info("shutdown signal received")
			cancel()
		}
	}()

	defer func() {
		signal.Stop(sig)
		close(sig)
	}()

	execute(ctx)
}
