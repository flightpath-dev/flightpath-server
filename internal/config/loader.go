package config

import (
	"log"
	"os"
	"strconv"
)

// Load loads configuration from environment variables
// Falls back to defaults for any missing values
func Load() *Config {
	cfg := Default()

	// Override with environment variables if present
	if port := os.Getenv("FLIGHTPATH_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}

	if host := os.Getenv("FLIGHTPATH_HOST"); host != "" {
		cfg.Server.Host = host
	}

	if logLevel := os.Getenv("FLIGHTPATH_LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	if mavPort := os.Getenv("FLIGHTPATH_MAVLINK_PORT"); mavPort != "" {
		cfg.MAVLink.DefaultPort = mavPort
	}

	if mavBaud := os.Getenv("FLIGHTPATH_MAVLINK_BAUD"); mavBaud != "" {
		if b, err := strconv.Atoi(mavBaud); err == nil {
			cfg.MAVLink.DefaultBaudRate = b
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return cfg
}
