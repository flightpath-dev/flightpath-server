package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	droneConnect "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1/dronev1connect"
	"github.com/flightpath-dev/flightpath-server/internal/config"
	"github.com/flightpath-dev/flightpath-server/internal/server"
	"github.com/flightpath-dev/flightpath-server/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create server
	srv := server.New(cfg)

	// Get shared dependencies
	deps := srv.GetDependencies()

	// Register services
	registerServices(srv, deps)

	// Setup graceful shutdown
	go handleShutdown(srv, deps)

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// registerServices registers all Connect services
func registerServices(srv *server.Server, deps *server.Dependencies) {
	// Connection service (fully implemented)
	connServer := services.NewConnectionServer(deps)
	connPath, connHandler := droneConnect.NewConnectionServiceHandler(connServer)
	srv.RegisterService(connPath, connHandler)

	// Control service (fully implemented)
	ctrlServer := services.NewControlServer(deps)
	ctrlPath, ctrlHandler := droneConnect.NewControlServiceHandler(ctrlServer)
	srv.RegisterService(ctrlPath, ctrlHandler)

	// Telemetry service (skeleton implementation)
	telemetryServer := services.NewTelemetryServer(deps)
	telemetryPath, telemetryHandler := droneConnect.NewTelemetryServiceHandler(telemetryServer)
	srv.RegisterService(telemetryPath, telemetryHandler)

	// Mission service (skeleton implementation)
	missionServer := services.NewMissionServer(deps)
	missionPath, missionHandler := droneConnect.NewMissionServiceHandler(missionServer)
	srv.RegisterService(missionPath, missionHandler)
}

// handleShutdown handles graceful shutdown on interrupt signals
func handleShutdown(srv *server.Server, deps *server.Dependencies) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Println("\n🛑 Shutting down server gracefully...")

	// Close MAVLink connection if exists
	if deps.HasMAVLinkClient() {
		client := deps.GetMAVLinkClient()
		if err := client.Close(); err != nil {
			log.Printf("Error closing MAVLink connection: %v", err)
		}
	}

	log.Println("✅ Cleanup complete")
	os.Exit(0)
}
