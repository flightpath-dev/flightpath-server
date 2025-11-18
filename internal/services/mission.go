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

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.UploadMissionResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Validate mission
	if len(req.Msg.Mission.Waypoints) == 0 {
		return connect.NewResponse(&drone.UploadMissionResponse{
			Success: false,
			Message: "Mission must have at least one waypoint",
		}), nil
	}

	// Upload mission via MAVLink
	err := client.UploadMission(req.Msg.Mission.Waypoints)
	if err != nil {
		return connect.NewResponse(&drone.UploadMissionResponse{
			Success: false,
			Message: fmt.Sprintf("Mission upload failed: %v", err),
		}), nil
	}

	logger.Printf("Mission uploaded successfully: %d waypoints", len(req.Msg.Mission.Waypoints))

	return connect.NewResponse(&drone.UploadMissionResponse{
		Success:           true,
		Message:           "Mission uploaded successfully",
		WaypointsUploaded: int32(len(req.Msg.Mission.Waypoints)),
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
	// This requires MISSION_REQUEST_LIST and handling MISSION_COUNT/MISSION_ITEM responses

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

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.StartMissionResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Set mission mode (AUTO with MISSION sub-mode)
	autoMissionMode := uint32(mavlink.PX4_MAIN_MODE_AUTO | (mavlink.PX4_AUTO_MODE_MISSION << 16))
	if err := client.SetMode(autoMissionMode); err != nil {
		return connect.NewResponse(&drone.StartMissionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to set AUTO mode: %v", err),
		}), nil
	}

	// Set current waypoint to 0 (start from beginning)
	if err := client.StartMission(0); err != nil {
		return connect.NewResponse(&drone.StartMissionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to start mission: %v", err),
		}), nil
	}

	logger.Println("Mission started successfully")

	return connect.NewResponse(&drone.StartMissionResponse{
		Success: true,
		Message: "Mission started successfully",
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

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.PauseMissionResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Switch to LOITER mode to pause (holds current position)
	autoLoiterMode := uint32(mavlink.PX4_MAIN_MODE_AUTO | (mavlink.PX4_AUTO_MODE_LOITER << 16))
	if err := client.SetMode(autoLoiterMode); err != nil {
		return connect.NewResponse(&drone.PauseMissionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to pause mission: %v", err),
		}), nil
	}

	logger.Println("Mission paused successfully")

	return connect.NewResponse(&drone.PauseMissionResponse{
		Success: true,
		Message: "Mission paused successfully",
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

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.ResumeMissionResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Switch back to AUTO MISSION mode
	autoMissionMode := uint32(mavlink.PX4_MAIN_MODE_AUTO | (mavlink.PX4_AUTO_MODE_MISSION << 16))
	if err := client.SetMode(autoMissionMode); err != nil {
		return connect.NewResponse(&drone.ResumeMissionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to resume mission: %v", err),
		}), nil
	}

	logger.Println("Mission resumed successfully")

	return connect.NewResponse(&drone.ResumeMissionResponse{
		Success: true,
		Message: "Mission resumed successfully",
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

	client := s.deps.GetMAVLinkClient()

	// Check if connected
	if !client.IsConnected() {
		return connect.NewResponse(&drone.ClearMissionResponse{
			Success: false,
			Message: "Drone is not connected",
		}), nil
	}

	// Clear mission via MAVLink
	if err := client.ClearMission(); err != nil {
		return connect.NewResponse(&drone.ClearMissionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to clear mission: %v", err),
		}), nil
	}

	logger.Println("Mission cleared successfully")

	return connect.NewResponse(&drone.ClearMissionResponse{
		Success: true,
		Message: "Mission cleared successfully",
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

	client := s.deps.GetMAVLinkClient()

	// Get mission progress from MAVLink client
	currentWaypoint, totalWaypoints, active := client.GetMissionProgress()

	var status drone.GetProgressResponse_Status
	if !active {
		status = drone.GetProgressResponse_STATUS_IDLE
	} else if currentWaypoint >= 0 && currentWaypoint < totalWaypoints {
		status = drone.GetProgressResponse_STATUS_IN_PROGRESS
	} else if currentWaypoint >= totalWaypoints {
		status = drone.GetProgressResponse_STATUS_COMPLETED
	} else {
		status = drone.GetProgressResponse_STATUS_IDLE
	}

	return connect.NewResponse(&drone.GetProgressResponse{
		Status:          status,
		CurrentWaypoint: currentWaypoint,
		TotalWaypoints:  totalWaypoints,
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

	client := s.deps.GetMAVLinkClient()

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
			// Get mission progress from MAVLink client
			currentWaypoint, totalWaypoints, active := client.GetMissionProgress()

			var status drone.StreamProgressResponse_Status
			if !active {
				status = drone.StreamProgressResponse_STATUS_IDLE
			} else if currentWaypoint >= 0 && currentWaypoint < totalWaypoints {
				status = drone.StreamProgressResponse_STATUS_IN_PROGRESS
			} else if currentWaypoint >= totalWaypoints {
				status = drone.StreamProgressResponse_STATUS_COMPLETED
			} else {
				status = drone.StreamProgressResponse_STATUS_IDLE
			}

			progress := &drone.StreamProgressResponse{
				Status:          status,
				CurrentWaypoint: currentWaypoint,
				TotalWaypoints:  totalWaypoints,
			}

			if err := stream.Send(progress); err != nil {
				logger.Printf("StreamProgress: Error sending: %v", err)
				return err
			}
		}
	}
}
