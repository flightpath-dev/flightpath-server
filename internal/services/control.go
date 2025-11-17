package services

import (
	"context"

	"connectrpc.com/connect"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
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

	// TODO: Implement flight mode change via MAVLink
	// This requires mapping generic FlightMode enum to MAVLink modes

	return connect.NewResponse(&drone.SetFlightModeResponse{
		Success: false,
		Message: "Flight mode change not yet implemented",
	}), nil
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

	return connect.NewResponse(&drone.GoToPositionResponse{
		Success: false,
		Message: "Go to position not yet implemented",
	}), nil
}

func (s *ControlServer) EmergencyStop(
	ctx context.Context,
	req *connect.Request[drone.EmergencyStopRequest],
) (*connect.Response[drone.EmergencyStopResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("⚠️  EMERGENCY STOP request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.EmergencyStopResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement emergency motor stop via MAVLink
	// This is MAV_CMD_DO_MOTOR_STOP or similar

	return connect.NewResponse(&drone.EmergencyStopResponse{
		Success: false,
		Message: "Emergency stop not yet implemented",
	}), nil
}
