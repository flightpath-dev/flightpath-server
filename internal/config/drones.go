package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DroneConfig represents a single drone's configuration
type DroneConfig struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Protocol    string                 `yaml:"protocol"` // "mavlink", "dji", etc.
	Connection  map[string]interface{} `yaml:"connection"`
}

// DroneRegistry holds all configured drones
type DroneRegistry struct {
	Drones []DroneConfig `yaml:"drones"`
}

// LoadDroneRegistry loads drone configurations from a YAML file
func LoadDroneRegistry(path string) (*DroneRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read drone registry: %w", err)
	}

	var registry DroneRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse drone registry: %w", err)
	}

	return &registry, nil
}

// FindDrone finds a drone by ID
func (r *DroneRegistry) FindDrone(id string) (*DroneConfig, error) {
	for _, drone := range r.Drones {
		if drone.ID == id {
			return &drone, nil
		}
	}
	return nil, fmt.Errorf("drone not found: %s", id)
}

// GetConnectionString returns a connection parameter as string
func (d *DroneConfig) GetConnectionString(key string) string {
	if val, ok := d.Connection[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetConnectionInt returns a connection parameter as int
func (d *DroneConfig) GetConnectionInt(key string) int {
	if val, ok := d.Connection[key]; ok {
		if num, ok := val.(int); ok {
			return num
		}
	}
	return 0
}
