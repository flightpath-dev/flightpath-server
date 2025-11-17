package services

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
	"github.com/flightpath-dev/flightpath-server/internal/server"
)

// MissionServer implements the MissionService
type MissionServer struct {
	deps *server.Dependencies
}

// NewMissionServer creates a new MissionServer
func NewMissionServer(deps *server.Dependencies) *MissionServer {
	return &MissionServer{
		deps: deps,
	}
}

// UploadMission uploads a mission to the drone
func (s *MissionServer) UploadMission(
	ctx context.Context,
	req *connect.Request[drone.UploadMissionRequest],
) (*connect.Response[drone.UploadMissionResponse], error) {
	logger := s.deps.GetLogger()
	logger.Printf("UploadMission request: mission_id=%s, waypoints=%d",
		req.Msg.Mission.Id, len(req.Msg.Mission.Waypoints))

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.UploadMissionResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement mission upload via MAVLink
	// This requires sending MISSION_COUNT, MISSION_ITEM_INT messages

	return connect.NewResponse(&drone.UploadMissionResponse{
		Success: false,
		Message: "Mission upload not yet implemented",
	}), nil
}

// DownloadMission downloads current mission from drone
func (s *MissionServer) DownloadMission(
	ctx context.Context,
	req *connect.Request[drone.DownloadMissionRequest],
) (*connect.Response[drone.DownloadMissionResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("DownloadMission request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.DownloadMissionResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement mission download via MAVLink

	return connect.NewResponse(&drone.DownloadMissionResponse{
		Success: false,
		Message: "Mission download not yet implemented",
	}), nil
}

// StartMission starts mission execution
func (s *MissionServer) StartMission(
	ctx context.Context,
	req *connect.Request[drone.StartMissionRequest],
) (*connect.Response[drone.StartMissionResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("StartMission request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.StartMissionResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement mission start via MAVLink
	// This requires setting mode to AUTO and sending MISSION_SET_CURRENT

	return connect.NewResponse(&drone.StartMissionResponse{
		Success: false,
		Message: "Mission start not yet implemented",
	}), nil
}

// PauseMission pauses mission execution
func (s *MissionServer) PauseMission(
	ctx context.Context,
	req *connect.Request[drone.PauseMissionRequest],
) (*connect.Response[drone.PauseMissionResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("PauseMission request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.PauseMissionResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement mission pause (usually via mode change to HOLD)

	return connect.NewResponse(&drone.PauseMissionResponse{
		Success: false,
		Message: "Mission pause not yet implemented",
	}), nil
}

// ResumeMission resumes mission execution
func (s *MissionServer) ResumeMission(
	ctx context.Context,
	req *connect.Request[drone.ResumeMissionRequest],
) (*connect.Response[drone.ResumeMissionResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("ResumeMission request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.ResumeMissionResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement mission resume (mode back to AUTO)

	return connect.NewResponse(&drone.ResumeMissionResponse{
		Success: false,
		Message: "Mission resume not yet implemented",
	}), nil
}

// ClearMission clears mission from drone
func (s *MissionServer) ClearMission(
	ctx context.Context,
	req *connect.Request[drone.ClearMissionRequest],
) (*connect.Response[drone.ClearMissionResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("ClearMission request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.ClearMissionResponse{
			Success: false,
			Message: "Not connected to drone",
		}), nil
	}

	// TODO: Implement mission clear (MISSION_CLEAR_ALL)

	return connect.NewResponse(&drone.ClearMissionResponse{
		Success: false,
		Message: "Mission clear not yet implemented",
	}), nil
}

// GetProgress gets current mission progress
func (s *MissionServer) GetProgress(
	ctx context.Context,
	req *connect.Request[drone.GetProgressRequest],
) (*connect.Response[drone.GetProgressResponse], error) {
	logger := s.deps.GetLogger()
	logger.Println("GetProgress request")

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewResponse(&drone.GetProgressResponse{
			Status: drone.GetProgressResponse_STATUS_IDLE,
		}), nil
	}

	// TODO: Get actual mission progress from MAVLink
	// Parse MISSION_CURRENT and MISSION_ITEM_REACHED messages

	return connect.NewResponse(&drone.GetProgressResponse{
		Status:          drone.GetProgressResponse_STATUS_IDLE,
		CurrentWaypoint: 0,
		TotalWaypoints:  0,
	}), nil
}

// StreamProgress streams mission progress updates
func (s *MissionServer) StreamProgress(
	ctx context.Context,
	req *connect.Request[drone.StreamProgressRequest],
	stream *connect.ServerStream[drone.StreamProgressResponse],
) error {
	logger := s.deps.GetLogger()
	logger.Printf("StreamProgress request: interval_ms=%d", req.Msg.IntervalMs)

	// Check if MAVLink client exists
	if !s.deps.HasMAVLinkClient() {
		return connect.NewError(connect.CodeFailedPrecondition,
			fmt.Errorf("not connected to drone"))
	}

	// Calculate interval
	interval := time.Second
	if req.Msg.IntervalMs > 0 {
		interval = time.Duration(req.Msg.IntervalMs) * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Println("StreamProgress: Client disconnected")
			return nil

		case <-ticker.C:
			// TODO: Get actual mission progress from MAVLink
			progress := &drone.StreamProgressResponse{
				Status:          drone.StreamProgressResponse_STATUS_IDLE,
				CurrentWaypoint: 0,
				TotalWaypoints:  0,
			}

			if err := stream.Send(progress); err != nil {
				logger.Printf("StreamProgress: Error sending: %v", err)
				return err
			}
		}
	}
}
