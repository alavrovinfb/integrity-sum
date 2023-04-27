package logger_test

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/logger"
)

func TestInit(t *testing.T) {
	t.Run("Set log level to info", func(t *testing.T) {
		l := logger.Init("")
		assert.Equal(t, logrus.InfoLevel, l.Level)
	})

	t.Run("Set log level to debug", func(t *testing.T) {
		l := logger.Init("debug")
		assert.Equal(t, logrus.DebugLevel, l.Level)
	})
}
