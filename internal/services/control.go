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
	s.deps.GetLogger().Println("Arm request")

	// TODO: Implement arm command in next iteration

	return connect.NewResponse(&drone.ArmResponse{
		Success: true,
		Message: "Arm logic not yet implemented",
	}), nil
}

func (s *ControlServer) Disarm(
	ctx context.Context,
	req *connect.Request[drone.DisarmRequest],
) (*connect.Response[drone.DisarmResponse], error) {
	s.deps.GetLogger().Println("Disarm request")

	// TODO: Implement disarm command in next iteration

	return connect.NewResponse(&drone.DisarmResponse{
		Success: true,
		Message: "Disarm logic not yet implemented",
	}), nil
}

func (s *ControlServer) SetMode(
	ctx context.Context,
	req *connect.Request[drone.SetModeRequest],
) (*connect.Response[drone.SetModeResponse], error) {
	s.deps.GetLogger().Println("SetMode request")

	// TODO: Implement set mode command in next iteration

	return connect.NewResponse(&drone.SetModeResponse{
		Success: true,
		Message: "SetMode logic not yet implemented",
	}), nil
}

func (s *ControlServer) Takeoff(
	ctx context.Context,
	req *connect.Request[drone.TakeoffRequest],
) (*connect.Response[drone.TakeoffResponse], error) {
	s.deps.GetLogger().Println("Takeoff request")

	// TODO: Implement takeoff command in next iteration

	return connect.NewResponse(&drone.TakeoffResponse{
		Success: true,
		Message: "Takeoff logic not yet implemented",
	}), nil
}

func (s *ControlServer) Land(
	ctx context.Context,
	req *connect.Request[drone.LandRequest],
) (*connect.Response[drone.LandResponse], error) {
	s.deps.GetLogger().Println("Land request")

	// TODO: Implement land command in next iteration

	return connect.NewResponse(&drone.LandResponse{
		Success: true,
		Message: "Land logic not yet implemented",
	}), nil
}

func (s *ControlServer) ReturnHome(
	ctx context.Context,
	req *connect.Request[drone.ReturnHomeRequest],
) (*connect.Response[drone.ReturnHomeResponse], error) {
	s.deps.GetLogger().Println("ReturnHome request")

	// TODO: Implement return home command in next iteration

	return connect.NewResponse(&drone.ReturnHomeResponse{
		Success: true,
		Message: "ReturnHome logic not yet implemented",
	}), nil
}

func (s *ControlServer) GoToPosition(
	ctx context.Context,
	req *connect.Request[drone.GoToPositionRequest],
) (*connect.Response[drone.GoToPositionResponse], error) {
	s.deps.GetLogger().Println("GoToPosition request")

	// TODO: Implement go to position command in next iteration

	return connect.NewResponse(&drone.GoToPositionResponse{
		Success: true,
		Message: "GoToPosition logic not yet implemented",
	}), nil
}

func (s *ControlServer) SetVelocity(
	ctx context.Context,
	req *connect.Request[drone.SetVelocityRequest],
) (*connect.Response[drone.SetVelocityResponse], error) {
	s.deps.GetLogger().Println("SetVelocity request")

	// TODO: Implement set velocity command in next iteration

	return connect.NewResponse(&drone.SetVelocityResponse{
		Success: true,
		Message: "SetVelocity logic not yet implemented",
	}), nil
}

func (s *ControlServer) SetHome(
	ctx context.Context,
	req *connect.Request[drone.SetHomeRequest],
) (*connect.Response[drone.SetHomeResponse], error) {
	s.deps.GetLogger().Println("SetHome request")

	// TODO: Implement set home command in next iteration

	return connect.NewResponse(&drone.SetHomeResponse{
		Success: true,
		Message: "SetHome logic not yet implemented",
	}), nil
}

func (s *ControlServer) EmergencyStop(
	ctx context.Context,
	req *connect.Request[drone.EmergencyStopRequest],
) (*connect.Response[drone.EmergencyStopResponse], error) {
	s.deps.GetLogger().Println("EmergencyStop request")

	// TODO: Implement emergency stop command in next iteration

	return connect.NewResponse(&drone.EmergencyStopResponse{
		Success: true,
		Message: "EmergencyStop logic not yet implemented",
	}), nil
}
