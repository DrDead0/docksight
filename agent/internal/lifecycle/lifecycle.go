package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"docksight-agent/internal/logger"
)

// Manager coordinates graceful startup/shutdown for the agent process.
type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

// New creates a lifecycle manager that listens for SIGINT and SIGTERM.
func New() *Manager {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return &Manager{ctx: ctx, cancel: cancel}
}

// Context returns the lifecycle context cancelled on interrupt.
func (m *Manager) Context() context.Context {
	return m.ctx
}

// Wait blocks until SIGINT/SIGTERM is received, then runs optional shutdown hooks.
func (m *Manager) Wait(onShutdown ...func()) {
	<-m.ctx.Done()
	logger.Info("shutdown signal received")
	m.Shutdown(onShutdown...)
}

// Shutdown runs shutdown hooks exactly once and cancels the lifecycle context.
func (m *Manager) Shutdown(hooks ...func()) {
	m.once.Do(func() {
		for _, hook := range hooks {
			if hook != nil {
				hook()
			}
		}
		m.cancel()
	})
}
