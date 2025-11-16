package services

import (
	"context"

	"connectrpc.com/connect"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
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
	s.deps.GetLogger().Printf("Connect request: drone_id=%s, timeout_ms=%d", req.Msg.DroneId, req.Msg.TimeoutMs)

	// TODO: Implement MAVLink connection in next iteration

	return connect.NewResponse(&drone.ConnectResponse{
		Success: true,
		Message: "Connection logic not yet implemented",
	}), nil
}

func (s *ConnectionServer) GetStatus(
	ctx context.Context,
	req *connect.Request[drone.GetStatusRequest],
) (*connect.Response[drone.GetStatusResponse], error) {
	s.deps.GetLogger().Println("GetStatus request")

	// TODO: Return actual status in next iteration

	return connect.NewResponse(&drone.GetStatusResponse{
		Connected:        false,
		UptimeMs:         0,
		LastMessageMs:    0,
		MessagesReceived: 0,
		MessagesSent:     0,
	}), nil
}

func (s *ConnectionServer) Disconnect(
	ctx context.Context,
	req *connect.Request[drone.DisconnectRequest],
) (*connect.Response[drone.DisconnectResponse], error) {
	s.deps.GetLogger().Println("Disconnect request")

	// TODO: Implement disconnect logic in next iteration

	return connect.NewResponse(&drone.DisconnectResponse{
		Success: true,
		Message: "Disconnect logic not yet implemented",
	}), nil
}

func (s *ConnectionServer) StreamHealth(
	ctx context.Context,
	req *connect.Request[drone.StreamHealthRequest],
	stream *connect.ServerStream[drone.StreamHealthResponse],
) error {
	s.deps.GetLogger().Println("StreamHealth request")

	// TODO: Implement health streaming in next iteration

	return nil
}
