package server

import (
	"log"
	"sync"

	"github.com/flightpath-dev/flightpath-server/internal/config"
	"github.com/flightpath-dev/flightpath-server/internal/mavlink"
)

// Dependencies holds all shared dependencies for services
type Dependencies struct {
	Config        *config.Config
	DroneRegistry *config.DroneRegistry
	Logger        *log.Logger
	MAVLinkClient *mavlink.Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewDependencies creates a new Dependencies instance
func NewDependencies(cfg *config.Config) *Dependencies {
	logger := log.New(log.Writer(), "[flightpath] ", log.LstdFlags|log.Lshortfile)

	// Try to load drone registry
	registryPath := cfg.Server.DroneRegistryPath
	if registryPath == "" {
		registryPath = "./data/config/drones.yaml"
	}

	registry, err := config.LoadDroneRegistry(registryPath)
	if err != nil {
		logger.Printf("Warning: Could not load drone registry: %v", err)
		// Create empty registry if file doesn't exist
		registry = &config.DroneRegistry{Drones: []config.DroneConfig{}}
	} else {
		logger.Printf("Loaded drone registry with %d drones", len(registry.Drones))
	}

	return &Dependencies{
		Config:        cfg,
		DroneRegistry: registry,
		Logger:        logger,
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

// SetMAVLinkClient sets the MAVLink client
func (d *Dependencies) SetMAVLinkClient(client *mavlink.Client) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.MAVLinkClient = client
}

// GetMAVLinkClient returns the MAVLink client (thread-safe)
func (d *Dependencies) GetMAVLinkClient() *mavlink.Client {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.MAVLinkClient
}

// HasMAVLinkClient returns true if MAVLink client is set
func (d *Dependencies) HasMAVLinkClient() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.MAVLinkClient != nil
}

// ClearMAVLinkClient removes the MAVLink client from dependencies
func (d *Dependencies) ClearMAVLinkClient() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.MAVLinkClient = nil
}

// GetDroneRegistry returns the drone registry (thread-safe)
func (d *Dependencies) GetDroneRegistry() *config.DroneRegistry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.DroneRegistry
}
