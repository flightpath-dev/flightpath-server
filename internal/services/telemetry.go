package services

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
	"github.com/flightpath-dev/flightpath-server/internal/server"
)

// TelemetryServer implements the TelemetryService
type TelemetryServer struct {
	deps *server.Dependencies
}

// NewTelemetryServer creates a new TelemetryServer
func NewTelemetryServer(deps *server.Dependencies) *TelemetryServer {
	return &TelemetryServer{
		deps: deps,
	}
}

// StreamTelemetry streams real-time telemetry data
func (s *TelemetryServer) StreamTelemetry(
	ctx context.Context,
	req *connect.Request[drone.StreamTelemetryRequest],
	stream *connect.ServerStream[drone.StreamTelemetryResponse],
) error {
	logger := s.deps.GetLogger()
	logger.Printf("StreamTelemetry request: rate_hz=%d", req.Msg.RateHz)

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("not connected to drone"))
	}

	client := s.deps.GetMAVLinkClient()

	// Calculate interval from rate
	interval := time.Second
	if req.Msg.RateHz > 0 {
		interval = time.Second / time.Duration(req.Msg.RateHz)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Println("StreamTelemetry: Client disconnected")
			return nil

		case <-ticker.C:
			// TODO: Get actual telemetry from MAVLink client
			// For now, send basic status

			telemetry := &drone.StreamTelemetryResponse{
				TimestampMs: time.Now().UnixMilli(),
				Armed:       client.IsArmed(),
				Mode:        drone.FlightMode_FLIGHT_MODE_MANUAL, // TODO: Get from MAVLink
				// Position, Velocity, Attitude, Battery, Health will be added
				// when we implement proper MAVLink telemetry parsing
			}

			if err := stream.Send(telemetry); err != nil {
				logger.Printf("StreamTelemetry: Error sending: %v", err)
				return err
			}
		}
	}
}

// GetSnapshot returns current telemetry snapshot
func (s *TelemetryServer) GetSnapshot(
	ctx context.Context,
	req *connect.Request[drone.GetSnapshotRequest],
) (*connect.Response[drone.GetSnapshotResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("GetSnapshot request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("not connected to drone"))
	}

	client := s.deps.GetMAVLinkClient()

	// TODO: Get actual telemetry from MAVLink client
	// For now, return basic status
	snapshot := &drone.GetSnapshotResponse{
		TimestampMs: time.Now().UnixMilli(),
		Armed:       client.IsArmed(),
		Mode:        drone.FlightMode_FLIGHT_MODE_MANUAL, // TODO: Get from MAVLink
		// Position, Velocity, Attitude, Battery, Health, HomePosition, Capabilities
		// will be added when we implement proper MAVLink telemetry parsing
	}

	return connect.NewResponse(snapshot), nil
}
