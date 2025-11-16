package server

import (
	"log"
	"sync"

	"github.com/flightpath-dev/flightpath-server/internal/config"
)

// Dependencies holds all shared dependencies for services
type Dependencies struct {
	Config *config.Config
	Logger *log.Logger

	// MAVLink client will be added here in next iteration
	// mavlinkClient *mavlink.Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewDependencies creates a new Dependencies instance
func NewDependencies(cfg *config.Config) *Dependencies {
	logger := log.New(log.Writer(), "[flightpath] ", log.LstdFlags|log.Lshortfile)

	return &Dependencies{
		Config: cfg,
		Logger: logger,
	}
}

// SetLogger allows updating the logger (useful for testing)
func (d *Dependencies) SetLogger(logger *log.Logger) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Logger = logger
}

// GetLogger returns the logger (thread-safe)
func (d *Dependencies) GetLogger() *log.Logger {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Logger
}
