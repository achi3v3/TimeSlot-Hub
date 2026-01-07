package closer

import (
	"context"
	"fmt"
	"time"
)

func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.shuttingDown {
		m.mu.Unlock()
		return fmt.Errorf("shutdown already in progress")
	}
	m.shuttingDown = true
	m.mu.Unlock()
	m.logger.Info("Starting graceful shutdown...")

	timeout := 30 * time.Second
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var shutdownErrs []error

	for i := len(m.gracefuls); i >= 0; i-- {
		g := m.gracefuls[i]
		m.logger.Infof("Shutting down %s...", g.Name())
		if err := g.Shutdown(shutdownCtx); err != nil {
			shutdownErrs = append(shutdownErrs, fmt.Errorf("failed to shutdown %s: %w", g.Name(), err))
			m.logger.Errorf("failed to shutdown %s: %v", g.Name(), err)
		} else {
			m.logger.Infof("%s shutdown completed", g.Name())
		}

	}

	closeCtx, cancelClose := context.WithTimeout(ctx, 10*time.Second)
	defer cancelClose()

	// Сортируем closers по заданному порядку
	orderedClosers := m.sortClosersByOrder()

	for _, closer := range orderedClosers {
		m.logger.Infof("Closing %s...", closer.Name())
		if err := closer.Close(closeCtx); err != nil {
			shutdownErrs = append(shutdownErrs,
				fmt.Errorf("failed to close %s: %w", closer.Name(), err))
			m.logger.Errorf("Failed to close %s: %v", closer.Name(), err)
		} else {
			m.logger.Infof("%s closed", closer.Name())
		}
	}

	m.logger.Info("Graceful shutdown completed")

	if len(shutdownErrs) > 0 {
		return fmt.Errorf("shutdown errors: %v", shutdownErrs)
	}
	return nil
}
