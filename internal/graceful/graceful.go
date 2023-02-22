package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func Execute(pCtx context.Context, execute func(context.Context)) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(pCtx)
	go func() {
		if _, ok := <-sig; ok {
			cancel()
		}
	}()

	defer func() {
		signal.Stop(sig)
		close(sig)
	}()

	execute(ctx)
}
