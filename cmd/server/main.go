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
	go handleShutdown(srv)

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// registerServices registers all Connect services
func registerServices(srv *server.Server, deps *server.Dependencies) {
	// Connection service
	connServer := services.NewConnectionServer(deps)
	connPath, connHandler := droneConnect.NewConnectionServiceHandler(connServer)
	srv.RegisterService(connPath, connHandler)

	// Control service
	ctrlServer := services.NewControlServer(deps)
	ctrlPath, ctrlHandler := droneConnect.NewControlServiceHandler(ctrlServer)
	srv.RegisterService(ctrlPath, ctrlHandler)

	// Future services can be registered here:
	// telemetryServer := services.NewTelemetryServer(deps)
	// telePath, teleHandler := droneConnect.NewTelemetryServiceHandler(telemetryServer)
	// srv.RegisterService(telePath, teleHandler)
}

// handleShutdown handles graceful shutdown on interrupt signals
func handleShutdown(srv *server.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Println("\n🛑 Shutting down server gracefully...")

	// TODO: Add cleanup logic here (close MAVLink connections, etc.)

	os.Exit(0)
}
