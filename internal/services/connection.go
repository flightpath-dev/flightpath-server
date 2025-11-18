package services

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
	"github.com/flightpath-dev/flightpath-server/internal/config"
	"github.com/flightpath-dev/flightpath-server/internal/mavlink"
	"github.com/flightpath-dev/flightpath-server/internal/server"
)

// ConnectionServer implements the ConnectionService
type ConnectionServer struct {
	deps *server.Dependencies
}

// NewConnectionServer creates a new ConnectionServer
func NewConnectionServer(deps *server.Dependencies) *ConnectionServer {
	return &ConnectionServer{
		deps: deps,
	}
}

func (s *ConnectionServer) Connect(
	ctx context.Context,
	req *connect.Request[drone.ConnectRequest],
) (*connect.Response[drone.ConnectResponse], error) {
	logger := s.deps.GetLogger()
	logger.Printf("Connect request: drone_id=%s", req.Msg.DroneId)

	// Require drone_id
	if req.Msg.DroneId == "" {
		return connect.NewResponse(&drone.ConnectResponse{
			Success: false,
			Message: "drone_id is required",
		}), nil
	}

	// Check if already connected
	if s.deps.HasMAVLinkClient() {
		client := s.deps.GetMAVLinkClient()
		if client.IsConnected() {
			return connect.NewResponse(&drone.ConnectResponse{
				Success: false,
				Message: "Already connected to a drone. Disconnect first.",
			}), nil
		}

		// Clean up old disconnected client
		client.Close()
	}

	// Look up drone in registry
	registry := s.deps.GetDroneRegistry()
	droneConfig, err := registry.FindDrone(req.Msg.DroneId)
	if err != nil {
		// Drone not found in registry
		return connect.NewResponse(&drone.ConnectResponse{
			Success: false,
			Message: fmt.Sprintf("Drone not found in registry: %s. Available drones: %v",
				req.Msg.DroneId, s.getAvailableDroneIDs()),
		}), nil
	}

	logger.Printf("Found drone in registry: %s (%s) using protocol: %s",
		droneConfig.ID, droneConfig.Name, droneConfig.Protocol)

	// Route to appropriate protocol handler
	switch droneConfig.Protocol {
	case "mavlink":
		return s.connectMAVLink(ctx, req, droneConfig)
	case "dji":
		// TODO: Implement DJI protocol
		return connect.NewResponse(&drone.ConnectResponse{
			Success: false,
			Message: "DJI protocol not yet implemented",
		}), nil
	default:
		return connect.NewResponse(&drone.ConnectResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown protocol: %s", droneConfig.Protocol),
		}), nil
	}
}

// connectMAVLink handles MAVLink protocol connections
func (s *ConnectionServer) connectMAVLink(
	ctx context.Context,
	req *connect.Request[drone.ConnectRequest],
	droneConfig *config.DroneConfig,
) (*connect.Response[drone.ConnectResponse], error) {
	logger := s.deps.GetLogger()

	// Extract MAVLink connection parameters from drone config
	port := droneConfig.GetConnectionString("port")
	baudRate := droneConfig.GetConnectionInt("baud_rate")

	if port == "" {
		port = s.deps.Config.MAVLink.DefaultPort
		logger.Printf("No port specified in config, using default: %s", port)
	}
	if baudRate == 0 {
		baudRate = s.deps.Config.MAVLink.DefaultBaudRate
		logger.Printf("No baud rate specified in config, using default: %d", baudRate)
	}

	logger.Printf("Connecting to MAVLink drone on %s at %d baud", port, baudRate)

	// Get timeout (use from request or default to 5 seconds)
	timeout := 5 * time.Second
	if req.Msg.TimeoutMs > 0 {
		timeout = time.Duration(req.Msg.TimeoutMs) * time.Millisecond
	}

	// Create MAVLink client
	client, err := mavlink.NewClient(mavlink.Config{
		Port:     port,
		BaudRate: baudRate,
		Logger:   logger,
	})
	if err != nil {
		return connect.NewResponse(&drone.ConnectResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create MAVLink connection: %v", err),
		}), nil
	}

	// Wait for heartbeat (with timeout)
	if err := client.WaitForConnection(timeout); err != nil {
		client.Close()
		return connect.NewResponse(&drone.ConnectResponse{
			Success: false,
			Message: fmt.Sprintf("Connection timeout: %v", err),
		}), nil
	}

	// Store client in dependencies
	s.deps.SetMAVLinkClient(client)

	logger.Printf("Successfully connected to drone %s (MAVLink System ID: %d)",
		droneConfig.ID, client.GetSystemID())

	return connect.NewResponse(&drone.ConnectResponse{
		Success:      true,
		Message:      fmt.Sprintf("Connected to %s (System ID: %d)", droneConfig.Name, client.GetSystemID()),
		DroneId:      droneConfig.ID,
		DroneName:    droneConfig.Name,
		Manufacturer: "PX4", // TODO: Get from AUTOPILOT_VERSION message
		Model:        droneConfig.Description,
		// TODO: Get capabilities from drone
	}), nil
}

// getAvailableDroneIDs returns list of configured drone IDs
func (s *ConnectionServer) getAvailableDroneIDs() []string {
	registry := s.deps.GetDroneRegistry()
	ids := make([]string, len(registry.Drones))
	for i, drone := range registry.Drones {
		ids[i] = drone.ID
	}
	return ids
}

func (s *ConnectionServer) GetStatus(
	ctx context.Context,
	req *connect.Request[drone.GetStatusRequest],
) (*connect.Response[drone.GetStatusResponse], error) {
	s.deps.GetLogger().Println("GetStatus request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.GetStatusResponse{
			Connected: false,
			Armed:     false,
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	return connect.NewResponse(&drone.GetStatusResponse{
		Connected: client.IsConnected(),
		Armed:     client.IsArmed(),
	}), nil
}

func (s *ConnectionServer) Disconnect(
	ctx context.Context,
	req *connect.Request[drone.DisconnectRequest],
) (*connect.Response[drone.DisconnectResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("Disconnect request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.DisconnectResponse{
			Success: false,
			Message: "Not connected to any drone",
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	// Close the connection
	if err := client.Close(); err != nil {
		return connect.NewResponse(&drone.DisconnectResponse{
			Success: false,
			Message: fmt.Sprintf("Error closing connection: %v", err),
		}), nil
	}

	// Remove client from dependencies after closing
	s.deps.ClearMAVLinkClient()

	logger.Println("Successfully disconnected from drone")

	return connect.NewResponse(&drone.DisconnectResponse{
		Success: true,
		Message: "Disconnected successfully",
	}), nil
}

func (s *ConnectionServer) ListDrones(
	ctx context.Context,
	req *connect.Request[drone.ListDronesRequest],
) (*connect.Response[drone.ListDronesResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("ListDrones request")

	registry := s.deps.GetDroneRegistry()
	drones := make([]*drone.DroneInfo, 0, len(registry.Drones))

	for _, droneConfig := range registry.Drones {
		drones = append(drones, &drone.DroneInfo{
			Id:          droneConfig.ID,
			Name:        droneConfig.Name,
			Description: droneConfig.Description,
			Protocol:    droneConfig.Protocol,
		})
	}

	return connect.NewResponse(&drone.ListDronesResponse{
		Drones: drones,
	}), nil
}
