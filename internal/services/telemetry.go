package services

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
	"github.com/flightpath-dev/flightpath-server/internal/mavlink"
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
			// Get telemetry from MAVLink client
			telemetry := client.GetTelemetry()

			response := &drone.StreamTelemetryResponse{
				TimestampMs: time.Now().UnixMilli(),

				// Position
				Position: &drone.Position{
					Latitude:  telemetry.Latitude,
					Longitude: telemetry.Longitude,
					Altitude:  telemetry.Altitude,
				},

				// Velocity
				Velocity: &drone.Velocity{
					X: telemetry.VelocityX,
					Y: telemetry.VelocityY,
					Z: telemetry.VelocityZ,
				},

				// Attitude
				Attitude: &drone.Attitude{
					Roll:  telemetry.Roll,
					Pitch: telemetry.Pitch,
					Yaw:   telemetry.Yaw,
				},

				// Battery
				Battery: &drone.BatteryStatus{
					Voltage:   telemetry.BatteryVoltage,
					Current:   telemetry.BatteryCurrent,
					Remaining: telemetry.BatteryRemaining,
				},

				// Health
				Health: &drone.SystemHealth{
					SensorsOk: telemetry.SensorsHealthy,
					GpsOk:     telemetry.SatelliteCount >= 6,
				},

				// Status
				Armed:         client.IsArmed(),
				Mode:          s.mapPX4ModeToFlightMode(telemetry.CustomMode),
				Heading:       telemetry.Heading,
				GroundSpeed:   telemetry.GroundSpeed,
				VerticalSpeed: telemetry.VerticalSpeed,

				// GPS
				GpsAccuracy:    telemetry.GPSAccuracy,
				SatelliteCount: telemetry.SatelliteCount,
			}

			if err := stream.Send(response); err != nil {
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
	telemetry := client.GetTelemetry()

	snapshot := &drone.GetSnapshotResponse{
		TimestampMs: time.Now().UnixMilli(),

		// Position
		Position: &drone.Position{
			Latitude:  telemetry.Latitude,
			Longitude: telemetry.Longitude,
			Altitude:  telemetry.Altitude,
		},

		// Velocity
		Velocity: &drone.Velocity{
			X: telemetry.VelocityX,
			Y: telemetry.VelocityY,
			Z: telemetry.VelocityZ,
		},

		// Attitude
		Attitude: &drone.Attitude{
			Roll:  telemetry.Roll,
			Pitch: telemetry.Pitch,
			Yaw:   telemetry.Yaw,
		},

		// Battery
		Battery: &drone.BatteryStatus{
			Voltage:   telemetry.BatteryVoltage,
			Current:   telemetry.BatteryCurrent,
			Remaining: telemetry.BatteryRemaining,
		},

		// Health
		Health: &drone.SystemHealth{
			SensorsOk: telemetry.SensorsHealthy,
			GpsOk:     telemetry.SatelliteCount >= 6,
		},

		// Status
		Armed: client.IsArmed(),
		Mode:  s.mapPX4ModeToFlightMode(telemetry.CustomMode),

		// Home position (will be zero until mission planning tracks it)
		HomePosition: &drone.Position{
			Latitude:  0,
			Longitude: 0,
			Altitude:  0,
		},

		// Capabilities
		Capabilities: &drone.Capabilities{
			HasGps:        telemetry.SatelliteCount > 0,
			HasCompass:    true,
			CanTakeoff:    true,
			CanLand:       true,
			CanReturnHome: true,
		},
	}

	return connect.NewResponse(snapshot), nil
}

// mapPX4ModeToFlightMode maps PX4 custom mode back to generic FlightMode
func (s *TelemetryServer) mapPX4ModeToFlightMode(customMode uint32) drone.FlightMode {
	// Extract main mode (lower 16 bits)
	mainMode := customMode & 0xFF

	// Extract sub mode (upper 16 bits)
	subMode := (customMode >> 16) & 0xFF

	// Map main modes
	switch mainMode {
	case mavlink.PX4_MAIN_MODE_MANUAL:
		return drone.FlightMode_FLIGHT_MODE_MANUAL

	case mavlink.PX4_MAIN_MODE_STABILIZED:
		return drone.FlightMode_FLIGHT_MODE_STABILIZED

	case mavlink.PX4_MAIN_MODE_ALTCTL:
		return drone.FlightMode_FLIGHT_MODE_ALTITUDE_HOLD

	case mavlink.PX4_MAIN_MODE_POSCTL:
		return drone.FlightMode_FLIGHT_MODE_POSITION_HOLD

	case mavlink.PX4_MAIN_MODE_OFFBOARD:
		return drone.FlightMode_FLIGHT_MODE_GUIDED

	case mavlink.PX4_MAIN_MODE_AUTO:
		// Map AUTO sub-modes
		switch subMode {
		case mavlink.PX4_AUTO_MODE_MISSION:
			return drone.FlightMode_FLIGHT_MODE_AUTO
		case mavlink.PX4_AUTO_MODE_RTL:
			return drone.FlightMode_FLIGHT_MODE_RETURN_HOME
		case mavlink.PX4_AUTO_MODE_LAND:
			return drone.FlightMode_FLIGHT_MODE_LAND
		case mavlink.PX4_AUTO_MODE_TAKEOFF:
			return drone.FlightMode_FLIGHT_MODE_TAKEOFF
		case mavlink.PX4_AUTO_MODE_LOITER:
			return drone.FlightMode_FLIGHT_MODE_LOITER
		default:
			return drone.FlightMode_FLIGHT_MODE_AUTO
		}

	default:
		return drone.FlightMode_FLIGHT_MODE_MANUAL
	}
}
