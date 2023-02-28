package integritymonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/core/ports"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/services/filehash"
	"github.com/ScienceSoft-Inc/integrity-sum/internal/utils/process"
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/alerts"
	"github.com/sirupsen/logrus"
)

type IntegrityMonitor struct {
	logger             *logrus.Logger
	fshasher           *filehash.FileSystemHasher
	repository         ports.IAppRepository
	kubeclient         ports.IKuberService
	alertSender        alerts.Sender
	delay              time.Duration
	monitorProcess     string
	monitorProcessPath string
}

func New(logger *logrus.Logger,
	fshasher *filehash.FileSystemHasher,
	repository ports.IAppRepository,
	kubeclient ports.IKuberService,
	alertSender alerts.Sender,
	delay time.Duration,
	monitorProcess string,
	monitorProcessPath string,

) *IntegrityMonitor {
	return &IntegrityMonitor{
		logger:             logger,
		fshasher:           fshasher,
		repository:         repository,
		kubeclient:         kubeclient,
		alertSender:        alertSender,
		delay:              delay,
		monitorProcess:     monitorProcess,
		monitorProcessPath: monitorProcessPath,
	}
}

func (m *IntegrityMonitor) Run(ctx context.Context) error {
	processPath, err := m.getProcessPath(m.monitorProcess, m.monitorProcessPath)
	if err != nil {
		return err
	}

	err = m.setupIntegrity(ctx, processPath)
	if err != nil {
		m.logger.WithError(err).Error("failed check integrity")
		return err
	}

	ticker := time.NewTicker(m.delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := m.processIntegrity(ctx, processPath)
			if err != nil {
				m.logger.WithError(err).Error("failed check integrity")
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *IntegrityMonitor) processIntegrity(ctx context.Context, path string) error {

	return m.checkIntegrity(ctx, path)
}

func (m *IntegrityMonitor) setupIntegrity(ctx context.Context, path string) error {

	return nil
}

func (m *IntegrityMonitor) checkIntegrity(ctx context.Context, path string) error {
	err := m.fshasher.CalculateInCallback(ctx, path, func(fh filehash.FileHash) error {

		return nil
	})
	return err
}

func (m *IntegrityMonitor) getProcessPath(procName string, path string) (string, error) {
	pid, err := process.GetPID(procName)
	if err != nil {
		return "", fmt.Errorf("failed build process path: %w", err)
	}
	return fmt.Sprintf("/proc/%d/root/%s", pid, path), nil
}
