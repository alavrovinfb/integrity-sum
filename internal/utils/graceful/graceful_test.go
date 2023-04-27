package graceful_test

import (
	"context"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/graceful"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
	"time"
)

// Testing whether passed function is executed and exits after given timeout or not
func TestExecute(t *testing.T) {
	logger := logrus.New()
	logger.Out = os.Stdout
	timeout := 1000 * time.Millisecond

	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	executed := false

	graceful.Execute(ctx, logger, func(ctx context.Context) {
		executed = true

		select {
		case <-ctx.Done():
			return
		}
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.True(t, executed, "execute function should be called")
		assert.True(t, time.Since(start) >= timeout, "execute function should be exited after timeout")
	}()
	wg.Wait()
}
