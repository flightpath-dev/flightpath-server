package services

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
	"github.com/flightpath-dev/flightpath-server/internal/mavlink"
	"github.com/flightpath-dev/flightpath-server/internal/server"
)

// ControlServer implements the ControlService
type ControlServer struct {
	deps *server.Dependencies
}

// NewControlServer creates a new ControlServer
func NewControlServer(deps *server.Dependencies) *ControlServer {
	return &ControlServer{
		deps: deps,
	}
}

func (s *ControlServer) Arm(
	ctx context.Context,
	req *connect.Request[drone.ArmRequest],
) (*connect.Response[drone.ArmResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("Arm request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.ArmResponse{
			Success: false,
			Message: "Not connected to drone. Call Connect first.",
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.ArmResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Send arm command
	if err := client.Arm(); err != nil {
		return connect.NewResponse(&drone.ArmResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	return connect.NewResponse(&drone.ArmResponse{
		Success: true,
		Message: "Arm command sent successfully",
	}), nil
}

func (s *ControlServer) Disarm(
	ctx context.Context,
	req *connect.Request[drone.DisarmRequest],
) (*connect.Response[drone.DisarmResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("Disarm request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.DisarmResponse{
			Success: false,
			Message: "Not connected to drone. Call Connect first.",
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.DisarmResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Send disarm command
	if err := client.Disarm(); err != nil {
		return connect.NewResponse(&drone.DisarmResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	return connect.NewResponse(&drone.DisarmResponse{
		Success: true,
		Message: "Disarm command sent successfully",
	}), nil
}

func (s *ControlServer) SetFlightMode(
	ctx context.Context,
	req *connect.Request[drone.SetFlightModeRequest],
) (*connect.Response[drone.SetFlightModeResponse], error) {
	logger := s.deps.GetLogger()
	logger.Printf("SetFlightMode request: mode=%s", req.Msg.Mode)

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.SetFlightModeResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.SetFlightModeResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Map generic FlightMode to PX4 custom mode
	customMode, err := s.mapFlightModeToPX4(req.Msg.Mode)
	if err != nil {
		return connect.NewResponse(&drone.SetFlightModeResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	// Send mode change command
	if err := client.SetMode(customMode); err != nil {
		return connect.NewResponse(&drone.SetFlightModeResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to set mode: %v", err),
		}), nil
	}

	logger.Printf("Successfully set mode to %s (PX4 custom mode: %d)", req.Msg.Mode, customMode)

	return connect.NewResponse(&drone.SetFlightModeResponse{
		Success:     true,
		Message:     fmt.Sprintf("Flight mode changed to %s", req.Msg.Mode),
		CurrentMode: req.Msg.Mode,
	}), nil
}

// mapFlightModeToPX4 maps generic FlightMode enum to PX4 custom mode
func (s *ControlServer) mapFlightModeToPX4(mode drone.FlightMode) (uint32, error) {
	switch mode {
	case drone.FlightMode_FLIGHT_MODE_MANUAL:
		// Manual mode - full manual control
		return mavlink.PX4_CUSTOM_MAIN_MODE_MANUAL, nil

	case drone.FlightMode_FLIGHT_MODE_STABILIZED:
		// Stabilized mode - attitude stabilization
		return mavlink.PX4_CUSTOM_MAIN_MODE_STABILIZED, nil

	case drone.FlightMode_FLIGHT_MODE_ALTITUDE_HOLD:
		// Altitude control mode
		return mavlink.PX4_CUSTOM_MAIN_MODE_ALTCTL, nil

	case drone.FlightMode_FLIGHT_MODE_POSITION_HOLD:
		// Position control mode (holds GPS position)
		return mavlink.PX4_CUSTOM_MAIN_MODE_POSCTL, nil

	case drone.FlightMode_FLIGHT_MODE_GUIDED:
		// Offboard/Guided mode (accepts external position commands)
		// In PX4, this is OFFBOARD mode
		return mavlink.PX4_CUSTOM_MAIN_MODE_OFFBOARD, nil

	case drone.FlightMode_FLIGHT_MODE_AUTO:
		// Auto mode - mission mode
		// Main mode AUTO + sub mode MISSION
		return s.encodePX4AutoMode(mavlink.PX4_CUSTOM_SUB_MODE_AUTO_MISSION), nil

	case drone.FlightMode_FLIGHT_MODE_RETURN_HOME:
		// Return to launch mode
		// Main mode AUTO + sub mode RTL
		return s.encodePX4AutoMode(mavlink.PX4_CUSTOM_SUB_MODE_AUTO_RTL), nil

	case drone.FlightMode_FLIGHT_MODE_LAND:
		// Land mode
		// Main mode AUTO + sub mode LAND
		return s.encodePX4AutoMode(mavlink.PX4_CUSTOM_SUB_MODE_AUTO_LAND), nil

	case drone.FlightMode_FLIGHT_MODE_TAKEOFF:
		// Takeoff mode
		// Main mode AUTO + sub mode TAKEOFF
		return s.encodePX4AutoMode(mavlink.PX4_CUSTOM_SUB_MODE_AUTO_TAKEOFF), nil

	case drone.FlightMode_FLIGHT_MODE_LOITER:
		// Loiter mode (circle around current position)
		// Main mode AUTO + sub mode LOITER
		return s.encodePX4AutoMode(mavlink.PX4_CUSTOM_SUB_MODE_AUTO_LOITER), nil

	default:
		return 0, fmt.Errorf("unsupported flight mode: %s", mode)
	}
}

// encodePX4AutoMode encodes PX4 AUTO main mode with sub mode
// PX4 custom mode format: main_mode | (sub_mode << 16)
func (s *ControlServer) encodePX4AutoMode(subMode uint32) uint32 {
	return mavlink.PX4_CUSTOM_MAIN_MODE_AUTO | (subMode << 16)
}

func (s *ControlServer) Takeoff(
	ctx context.Context,
	req *connect.Request[drone.TakeoffRequest],
) (*connect.Response[drone.TakeoffResponse], error) {
	logger := s.deps.GetLogger()
	logger.Printf("Takeoff request: altitude=%.2fm", req.Msg.Altitude)

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.TakeoffResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.TakeoffResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Send takeoff command
	if err := client.Takeoff(float32(req.Msg.Altitude)); err != nil {
		return connect.NewResponse(&drone.TakeoffResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	return connect.NewResponse(&drone.TakeoffResponse{
		Success: true,
		Message: "Takeoff command sent successfully",
	}), nil
}

func (s *ControlServer) Land(
	ctx context.Context,
	req *connect.Request[drone.LandRequest],
) (*connect.Response[drone.LandResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("Land request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.LandResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.LandResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Send land command
	if err := client.Land(); err != nil {
		return connect.NewResponse(&drone.LandResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	return connect.NewResponse(&drone.LandResponse{
		Success: true,
		Message: "Land command sent successfully",
	}), nil
}

func (s *ControlServer) ReturnHome(
	ctx context.Context,
	req *connect.Request[drone.ReturnHomeRequest],
) (*connect.Response[drone.ReturnHomeResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("ReturnHome request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.ReturnHomeResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.ReturnHomeResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Send return to launch command
	if err := client.ReturnToLaunch(); err != nil {
		return connect.NewResponse(&drone.ReturnHomeResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	return connect.NewResponse(&drone.ReturnHomeResponse{
		Success: true,
		Message: "Return home command sent successfully",
	}), nil
}

func (s *ControlServer) GoToPosition(
	ctx context.Context,
	req *connect.Request[drone.GoToPositionRequest],
) (*connect.Response[drone.GoToPositionResponse], error) {
	logger := s.deps.GetLogger()
	logger.Printf("GoToPosition request: lat=%.6f, lon=%.6f, alt=%.2f",
		req.Msg.Target.Latitude, req.Msg.Target.Longitude, req.Msg.Target.Altitude)

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.GoToPositionResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement goto position via MAVLink
	// This requires SET_POSITION_TARGET_GLOBAL_INT message
	// Must be in GUIDED/OFFBOARD mode first

	return connect.NewResponse(&drone.GoToPositionResponse{
		Success: false,
		Message: "Go to position not yet implemented",
	}), nil
}
