package config

import (
	"fmt"
)

// Config holds all application configuration
type Config struct {
	Server  ServerConfig
	MAVLink MAVLinkConfig
	Logging LoggingConfig
}

type ServerConfig struct {
	Host              string
	Port              int
	CORSOrigins       []string
	DroneRegistryPath string // Path to drones.yaml
}

type MAVLinkConfig struct {
	// Default connection settings (can be overridden per drone)
	DefaultPort     string
	DefaultBaudRate int
}

type LoggingConfig struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "json", "text"
}

// Default returns a Config with sensible defaults
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			CORSOrigins: []string{
				"http://localhost:5173", // Vite dev server
				"http://localhost:3000",
			},
			DroneRegistryPath: "./data/config/drones.yaml",
		},
		MAVLink: MAVLinkConfig{
			DefaultPort:     "/dev/ttyUSB0",
			DefaultBaudRate: 57600,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	return nil
}

// ServerAddr returns the server address as host:port
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
