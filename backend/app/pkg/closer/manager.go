package closer

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

type Manager struct {
	mu            sync.RWMutex
	closers       []Closer
	gracefuls     []Graceful
	logger        *logrus.Logger
	shutdownOrder []string
	shuttingDown  bool
}

func NewManager(logger *logrus.Logger) *Manager {
	if logger == nil {
		logger = logrus.New()
	}
	return &Manager{
		logger:        logger,
		shutdownOrder: []string{"server", "reminder-scheduler", "database"},
	}
}

func (m *Manager) AddCloser(closer Closer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closers = append(m.closers, closer)
}

func (m *Manager) AddGraceful(graceful Graceful) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gracefuls = append(m.gracefuls, graceful)
}

func (m *Manager) WaitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	sig := <-sigChan
	m.logger.Infof("Received signal: %v", sig)
	m.Shutdown(context.Background())
}
