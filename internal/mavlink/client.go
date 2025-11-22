package mavlink

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/message"

	drone "github.com/flightpath-dev/flightpath-proto/gen/go/drone/v1"
)

// PX4 Main Flight Modes
// These are standard PX4 modes encoded in MAVLink's custom_mode field
const (
	PX4_MAIN_MODE_MANUAL     = 1
	PX4_MAIN_MODE_ALTCTL     = 2
	PX4_MAIN_MODE_POSCTL     = 3
	PX4_MAIN_MODE_AUTO       = 4
	PX4_MAIN_MODE_ACRO       = 5
	PX4_MAIN_MODE_OFFBOARD   = 6
	PX4_MAIN_MODE_STABILIZED = 7
	PX4_MAIN_MODE_RATTITUDE  = 8
)

// PX4 AUTO Sub-Modes
// When main mode is AUTO, these specify the AUTO behavior
const (
	PX4_AUTO_MODE_READY    = 1
	PX4_AUTO_MODE_TAKEOFF  = 2
	PX4_AUTO_MODE_LOITER   = 3
	PX4_AUTO_MODE_MISSION  = 4
	PX4_AUTO_MODE_RTL      = 5
	PX4_AUTO_MODE_LAND     = 6
	PX4_AUTO_MODE_FOLLOW   = 8
	PX4_AUTO_MODE_PRECLAND = 9
)

// Position target type mask bits
// These bits tell the autopilot which fields to use/ignore
const (
	// Position
	POSITION_TARGET_TYPEMASK_X_IGNORE = 0b0000000000000001
	POSITION_TARGET_TYPEMASK_Y_IGNORE = 0b0000000000000010
	POSITION_TARGET_TYPEMASK_Z_IGNORE = 0b0000000000000100
	// Velocity
	POSITION_TARGET_TYPEMASK_VX_IGNORE = 0b0000000000001000
	POSITION_TARGET_TYPEMASK_VY_IGNORE = 0b0000000000010000
	POSITION_TARGET_TYPEMASK_VZ_IGNORE = 0b0000000000100000
	// Acceleration
	POSITION_TARGET_TYPEMASK_AX_IGNORE = 0b0000000001000000
	POSITION_TARGET_TYPEMASK_AY_IGNORE = 0b0000000010000000
	POSITION_TARGET_TYPEMASK_AZ_IGNORE = 0b0000000100000000
	// Force/Yaw
	POSITION_TARGET_TYPEMASK_FORCE_SET       = 0b0000001000000000
	POSITION_TARGET_TYPEMASK_YAW_IGNORE      = 0b0000010000000000
	POSITION_TARGET_TYPEMASK_YAW_RATE_IGNORE = 0b0000100000000000
)

// TelemetryData holds current telemetry state
type TelemetryData struct {
	// Position (from GLOBAL_POSITION_INT)
	Latitude  float64 // degrees
	Longitude float64 // degrees
	Altitude  float64 // meters (MSL)

	// Velocity (from GLOBAL_POSITION_INT)
	VelocityX float64 // m/s (north)
	VelocityY float64 // m/s (east)
	VelocityZ float64 // m/s (down)

	// Attitude (from ATTITUDE)
	Roll  float64 // radians
	Pitch float64 // radians
	Yaw   float64 // radians

	// Navigation (from VFR_HUD)
	Heading       float64 // degrees
	GroundSpeed   float64 // m/s
	VerticalSpeed float64 // m/s

	// Battery (from SYS_STATUS)
	BatteryVoltage   float64 // volts
	BatteryRemaining int32   // percent
	BatteryCurrent   float64 // amps

	// GPS (from GPS_RAW_INT)
	GPSAccuracy    float64 // meters
	SatelliteCount int32

	// System health (from SYS_STATUS)
	SensorsHealthy bool

	// Flight mode (from HEARTBEAT)
	CustomMode uint32
	BaseMode   uint8

	// Timestamps
	LastUpdate time.Time
}

// MissionState holds mission upload/download state
type MissionState struct {
	Uploading        bool
	Downloading      bool
	Waypoints        []*drone.Waypoint
	CurrentIndex     int
	TotalCount       int
	UploadComplete   chan error
	DownloadComplete chan error

	// Mission progress
	CurrentWaypoint int32
	TotalWaypoints  int32
	MissionActive   bool
}

// Client represents a MAVLink connection to a drone
type Client struct {
	node      *gomavlib.Node
	systemID  uint8
	connected bool
	armed     bool
	logger    *log.Logger

	// Thread-safe state
	mu sync.RWMutex

	// Last heartbeat time
	lastHeartbeat time.Time

	// Connection parameters
	port     string
	baudRate int

	// Telemetry data
	telemetry TelemetryData

	// Mission state
	missionState MissionState

	// Ground station heartbeat
	stopHeartbeat chan struct{}
	heartbeatDone chan struct{}
}

// Config holds MAVLink client configuration
type Config struct {
	Port     string
	BaudRate int
	Logger   *log.Logger
}

// NewClient creates a new MAVLink client
func NewClient(cfg Config) (*Client, error) {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointSerial{
				Device: cfg.Port,
				Baud:   cfg.BaudRate,
			},
		},
		Dialect:     common.Dialect,
		OutVersion:  gomavlib.V2,
		OutSystemID: 255, // GCS system ID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MAVLink node: %w", err)
	}

	client := &Client{
		node:      node,
		logger:    cfg.Logger,
		connected: false,
		port:      cfg.Port,
		baudRate:  cfg.BaudRate,
		telemetry: TelemetryData{
			LastUpdate: time.Now(),
		},
		missionState:  MissionState{},
		stopHeartbeat: make(chan struct{}),
		heartbeatDone: make(chan struct{}),
	}

	// Start listening for messages
	go client.listen()

	// Start sending ground station heartbeat and system time
	go client.sendGroundStationMessages()

	return client, nil
}

// sendGroundStationMessages sends periodic HEARTBEAT and SYSTEM_TIME messages
// This identifies Flightpath as a ground station and provides GPS assistance
func (c *Client) sendGroundStationMessages() {
	defer close(c.heartbeatDone)
	c.logger.Println("MAVLink: Starting ground station message sender")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopHeartbeat:
			c.logger.Println("MAVLink: Stopping ground station message sender")
			return

		case <-ticker.C:
			// Send HEARTBEAT - identifies us as a ground control station
			// This satisfies PX4's COM_DL_LOSS_T requirement
			err := c.node.WriteMessageAll(&common.MessageHeartbeat{
				Type:           common.MAV_TYPE_GCS, // Ground Control Station
				Autopilot:      common.MAV_AUTOPILOT_INVALID,
				BaseMode:       0,
				CustomMode:     0,
				SystemStatus:   common.MAV_STATE_ACTIVE,
				MavlinkVersion: 3,
			})
			if err != nil {
				c.logger.Printf("MAVLink: Error sending HEARTBEAT: %v", err)
			}

			// Send SYSTEM_TIME - provides accurate time for GPS assistance
			// This helps GPS achieve lock faster (warm start vs cold start)
			currentTime := time.Now()
			err = c.node.WriteMessageAll(&common.MessageSystemTime{
				TimeUnixUsec: uint64(currentTime.UnixMicro()),
				TimeBootMs:   uint32(currentTime.UnixMilli() % (1 << 32)),
			})
			if err != nil {
				c.logger.Printf("MAVLink: Error sending SYSTEM_TIME: %v", err)
			}
		}
	}
}

// requestDataStreams requests telemetry data streams from the drone
// This ensures we receive regular updates of position, attitude, etc.
func (c *Client) requestDataStreams() error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	c.logger.Println("MAVLink: Requesting data streams from drone")

	// Request all data streams at 10 Hz
	return c.node.WriteMessageAll(&common.MessageRequestDataStream{
		TargetSystem:    systemID,
		TargetComponent: 1,
		ReqStreamId:     uint8(common.MAV_DATA_STREAM_ALL),
		ReqMessageRate:  10, // 10 Hz
		StartStop:       1,  // Start streaming
	})
}

// listen processes incoming MAVLink messages
func (c *Client) listen() {
	c.logger.Println("MAVLink: Starting message listener")

	for evt := range c.node.Events() {
		if frm, ok := evt.(*gomavlib.EventFrame); ok {
			c.handleMessage(frm.Message(), frm.SystemID(), frm.ComponentID())
		}
	}

	c.logger.Println("MAVLink: Message listener stopped")
}

// handleMessage processes individual MAVLink messages
func (c *Client) handleMessage(msg message.Message, sysID, compID uint8) {
	switch m := msg.(type) {
	case *common.MessageHeartbeat:
		c.handleHeartbeat(m, sysID)

	case *common.MessageCommandAck:
		c.handleCommandAck(m)

	case *common.MessageStatustext:
		c.logger.Printf("MAVLink STATUS: [%d] %s", m.Severity, m.Text)

	case *common.MessageGlobalPositionInt:
		c.handleGlobalPosition(m)

	case *common.MessageAttitude:
		c.handleAttitude(m)

	case *common.MessageVfrHud:
		c.handleVfrHud(m)

	case *common.MessageSysStatus:
		c.handleSysStatus(m)

	case *common.MessageGpsRawInt:
		c.handleGpsRaw(m)

	case *common.MessageMissionRequest:
		c.handleMissionRequest(m)

	case *common.MessageMissionRequestInt:
		c.handleMissionRequestInt(m)

	case *common.MessageMissionAck:
		c.handleMissionAck(m)

	case *common.MessageMissionCurrent:
		c.handleMissionCurrent(m)

	case *common.MessageMissionItemReached:
		c.handleMissionItemReached(m)
	}
}

// handleHeartbeat processes HEARTBEAT messages
func (c *Client) handleHeartbeat(msg *common.MessageHeartbeat, sysID uint8) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		c.logger.Printf("MAVLink: Connected to system %d", sysID)
	}

	c.connected = true
	c.systemID = sysID
	c.lastHeartbeat = time.Now()

	// Check armed status (bit 7 of base_mode)
	wasArmed := c.armed
	c.armed = (msg.BaseMode & common.MAV_MODE_FLAG_SAFETY_ARMED) != 0

	if wasArmed != c.armed {
		c.logger.Printf("MAVLink: Armed status changed: %v", c.armed)
	}

	// Store flight mode
	c.telemetry.CustomMode = msg.CustomMode
	c.telemetry.BaseMode = uint8(msg.BaseMode)
}

// handleGlobalPosition processes GLOBAL_POSITION_INT messages
func (c *Client) handleGlobalPosition(msg *common.MessageGlobalPositionInt) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Convert from 1E7 degrees to degrees
	c.telemetry.Latitude = float64(msg.Lat) / 1e7
	c.telemetry.Longitude = float64(msg.Lon) / 1e7

	// Convert from mm to meters
	c.telemetry.Altitude = float64(msg.Alt) / 1000.0

	// Convert from cm/s to m/s
	c.telemetry.VelocityX = float64(msg.Vx) / 100.0
	c.telemetry.VelocityY = float64(msg.Vy) / 100.0
	c.telemetry.VelocityZ = float64(msg.Vz) / 100.0

	c.telemetry.LastUpdate = time.Now()
}

// handleAttitude processes ATTITUDE messages
func (c *Client) handleAttitude(msg *common.MessageAttitude) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.telemetry.Roll = float64(msg.Roll)
	c.telemetry.Pitch = float64(msg.Pitch)
	c.telemetry.Yaw = float64(msg.Yaw)

	c.telemetry.LastUpdate = time.Now()
}

// handleVfrHud processes VFR_HUD messages
func (c *Client) handleVfrHud(msg *common.MessageVfrHud) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.telemetry.Heading = float64(msg.Heading)
	c.telemetry.GroundSpeed = float64(msg.Groundspeed)
	c.telemetry.VerticalSpeed = float64(msg.Climb)

	c.telemetry.LastUpdate = time.Now()
}

// handleSysStatus processes SYS_STATUS messages
func (c *Client) handleSysStatus(msg *common.MessageSysStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Convert from millivolts to volts
	c.telemetry.BatteryVoltage = float64(msg.VoltageBattery) / 1000.0
	c.telemetry.BatteryRemaining = int32(msg.BatteryRemaining)

	// Convert from centiamps to amps
	c.telemetry.BatteryCurrent = float64(msg.CurrentBattery) / 100.0

	// Check if critical sensors are healthy
	c.telemetry.SensorsHealthy = (msg.OnboardControlSensorsHealth &
		msg.OnboardControlSensorsEnabled) == msg.OnboardControlSensorsEnabled

	c.telemetry.LastUpdate = time.Now()
}

// handleGpsRaw processes GPS_RAW_INT messages
func (c *Client) handleGpsRaw(msg *common.MessageGpsRawInt) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// EPH (HDOP * 100) - convert to meters (approximate)
	c.telemetry.GPSAccuracy = float64(msg.Eph) / 100.0
	c.telemetry.SatelliteCount = int32(msg.SatellitesVisible)

	c.telemetry.LastUpdate = time.Now()
}

// handleMissionRequest processes MISSION_REQUEST messages
func (c *Client) handleMissionRequest(msg *common.MessageMissionRequest) {
	c.handleMissionRequestInt(&common.MessageMissionRequestInt{
		Seq: msg.Seq,
	})
}

// handleMissionRequestInt processes MISSION_REQUEST_INT messages
func (c *Client) handleMissionRequestInt(msg *common.MessageMissionRequestInt) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.missionState.Uploading {
		c.logger.Printf("MAVLink: Received unexpected MISSION_REQUEST_INT for seq %d", msg.Seq)
		return
	}

	seq := int(msg.Seq)
	if seq >= len(c.missionState.Waypoints) {
		c.logger.Printf("MAVLink: Invalid waypoint sequence %d (max %d)", seq, len(c.missionState.Waypoints))
		return
	}

	c.logger.Printf("MAVLink: Sending waypoint %d/%d", seq+1, len(c.missionState.Waypoints))

	// Send the requested waypoint
	wp := c.missionState.Waypoints[seq]
	if err := c.sendMissionItem(wp); err != nil {
		c.logger.Printf("MAVLink: Error sending waypoint %d: %v", seq, err)
		if c.missionState.UploadComplete != nil {
			c.missionState.UploadComplete <- err
			c.missionState.UploadComplete = nil
		}
		c.missionState.Uploading = false
	}
}

// handleMissionAck processes MISSION_ACK messages
func (c *Client) handleMissionAck(msg *common.MessageMissionAck) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Printf("MAVLink: Mission ACK received: type=%d", msg.Type)

	if c.missionState.Uploading {
		c.missionState.Uploading = false
		if c.missionState.UploadComplete != nil {
			if msg.Type == common.MAV_MISSION_ACCEPTED {
				c.logger.Println("MAVLink: Mission upload successful")
				c.missionState.UploadComplete <- nil
			} else {
				c.logger.Printf("MAVLink: Mission upload failed: %d", msg.Type)
				c.missionState.UploadComplete <- fmt.Errorf("mission upload failed: %d", msg.Type)
			}
			c.missionState.UploadComplete = nil
		}
	}
}

// handleMissionCurrent processes MISSION_CURRENT messages
func (c *Client) handleMissionCurrent(msg *common.MessageMissionCurrent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.missionState.CurrentWaypoint = int32(msg.Seq)
	c.missionState.MissionActive = msg.Seq >= 0

	c.logger.Printf("MAVLink: Current mission waypoint: %d", msg.Seq)
}

// handleMissionItemReached processes MISSION_ITEM_REACHED messages
func (c *Client) handleMissionItemReached(msg *common.MessageMissionItemReached) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Printf("MAVLink: Mission waypoint %d reached", msg.Seq)
}

// handleCommandAck processes command acknowledgments
func (c *Client) handleCommandAck(msg *common.MessageCommandAck) {
	result := "UNKNOWN"
	switch msg.Result {
	case common.MAV_RESULT_ACCEPTED:
		result = "ACCEPTED"
	case common.MAV_RESULT_TEMPORARILY_REJECTED:
		result = "TEMPORARILY_REJECTED"
	case common.MAV_RESULT_DENIED:
		result = "DENIED"
	case common.MAV_RESULT_UNSUPPORTED:
		result = "UNSUPPORTED"
	case common.MAV_RESULT_FAILED:
		result = "FAILED"
	case common.MAV_RESULT_IN_PROGRESS:
		result = "IN_PROGRESS"
	}

	c.logger.Printf("MAVLink: Command %d result: %s", msg.Command, result)
}

// GoToPosition sends a position setpoint to the drone
// The drone must be in GUIDED (OFFBOARD) mode to accept position commands
func (c *Client) GoToPosition(latitude, longitude, altitude float64) error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Printf("MAVLink: Sending position setpoint: lat=%.6f, lon=%.6f, alt=%.2f",
		latitude, longitude, altitude)

	// Convert to MAVLink format
	lat := int32(latitude * 1e7)  // degrees * 1E7
	lon := int32(longitude * 1e7) // degrees * 1E7
	alt := float32(altitude)      // meters MSL

	// Type mask: use only position (ignore velocity, acceleration, yaw)
	typeMask := uint16(
		POSITION_TARGET_TYPEMASK_VX_IGNORE |
			POSITION_TARGET_TYPEMASK_VY_IGNORE |
			POSITION_TARGET_TYPEMASK_VZ_IGNORE |
			POSITION_TARGET_TYPEMASK_AX_IGNORE |
			POSITION_TARGET_TYPEMASK_AY_IGNORE |
			POSITION_TARGET_TYPEMASK_AZ_IGNORE |
			POSITION_TARGET_TYPEMASK_YAW_IGNORE |
			POSITION_TARGET_TYPEMASK_YAW_RATE_IGNORE,
	)

	// Send SET_POSITION_TARGET_GLOBAL_INT message
	return c.node.WriteMessageAll(&common.MessageSetPositionTargetGlobalInt{
		TargetSystem:    systemID,
		TargetComponent: 1,
		TimeBootMs:      uint32(time.Now().UnixMilli()),
		CoordinateFrame: common.MAV_FRAME_GLOBAL_RELATIVE_ALT_INT,
		TypeMask:        common.POSITION_TARGET_TYPEMASK(typeMask),
		LatInt:          lat,
		LonInt:          lon,
		Alt:             alt,
		Vx:              0,
		Vy:              0,
		Vz:              0,
		Afx:             0,
		Afy:             0,
		Afz:             0,
		Yaw:             0,
		YawRate:         0,
	})
}

// UploadMission uploads a mission to the drone
func (c *Client) UploadMission(waypoints []*drone.Waypoint) error {
	c.mu.Lock()

	if c.missionState.Uploading {
		c.mu.Unlock()
		return fmt.Errorf("mission upload already in progress")
	}

	systemID := c.systemID
	c.missionState.Uploading = true
	c.missionState.Waypoints = waypoints
	c.missionState.TotalCount = len(waypoints)
	c.missionState.CurrentIndex = 0
	c.missionState.UploadComplete = make(chan error, 1)

	uploadComplete := c.missionState.UploadComplete
	c.mu.Unlock()

	c.logger.Printf("MAVLink: Starting mission upload (%d waypoints)", len(waypoints))

	// Send MISSION_COUNT
	err := c.node.WriteMessageAll(&common.MessageMissionCount{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Count:           uint16(len(waypoints)),
	})

	if err != nil {
		c.mu.Lock()
		c.missionState.Uploading = false
		c.mu.Unlock()
		return fmt.Errorf("failed to send MISSION_COUNT: %w", err)
	}

	// Wait for upload to complete (with timeout)
	select {
	case err := <-uploadComplete:
		return err
	case <-time.After(30 * time.Second):
		c.mu.Lock()
		c.missionState.Uploading = false
		c.mu.Unlock()
		return fmt.Errorf("mission upload timeout")
	}
}

// sendMissionItem sends a single mission item to the drone
func (c *Client) sendMissionItem(wp *drone.Waypoint) error {
	systemID := c.systemID

	// Map action to MAVLink command
	command := c.mapWaypointActionToMAVLink(wp.Action)

	// Convert position
	lat := int32(wp.Position.Latitude * 1e7)
	lon := int32(wp.Position.Longitude * 1e7)
	alt := float32(wp.Position.Altitude)

	return c.node.WriteMessageAll(&common.MessageMissionItemInt{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Seq:             uint16(wp.Sequence),
		Frame:           common.MAV_FRAME_GLOBAL_RELATIVE_ALT,
		Command:         command,
		Current:         0,
		Autocontinue:    1,
		Param1:          float32(wp.HoldTimeSec),
		Param2:          float32(wp.AcceptanceRadius),
		Param3:          0,
		Param4:          float32(wp.Heading),
		X:               lat,
		Y:               lon,
		Z:               alt,
	})
}

// mapWaypointActionToMAVLink maps proto waypoint action to MAVLink command
func (c *Client) mapWaypointActionToMAVLink(action drone.Waypoint_Action) common.MAV_CMD {
	switch action {
	case drone.Waypoint_ACTION_TAKEOFF:
		return common.MAV_CMD_NAV_TAKEOFF
	case drone.Waypoint_ACTION_LAND:
		return common.MAV_CMD_NAV_LAND
	case drone.Waypoint_ACTION_WAYPOINT:
		return common.MAV_CMD_NAV_WAYPOINT
	case drone.Waypoint_ACTION_LOITER:
		return common.MAV_CMD_NAV_LOITER_UNLIM
	case drone.Waypoint_ACTION_HOLD:
		return common.MAV_CMD_NAV_LOITER_TIME
	default:
		return common.MAV_CMD_NAV_WAYPOINT
	}
}

// ClearMission clears the mission from the drone
func (c *Client) ClearMission() error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Println("MAVLink: Clearing mission")

	return c.node.WriteMessageAll(&common.MessageMissionClearAll{
		TargetSystem:    systemID,
		TargetComponent: 1,
	})
}

// StartMission starts mission execution at specified waypoint
func (c *Client) StartMission(waypointIndex int32) error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Printf("MAVLink: Starting mission at waypoint %d", waypointIndex)

	// Send MISSION_SET_CURRENT
	return c.node.WriteMessageAll(&common.MessageMissionSetCurrent{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Seq:             uint16(waypointIndex),
	})
}

// GetMissionProgress returns current mission progress
func (c *Client) GetMissionProgress() (currentWaypoint int32, totalWaypoints int32, active bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.missionState.CurrentWaypoint, c.missionState.TotalWaypoints, c.missionState.MissionActive
}

// GetTelemetry returns current telemetry data (thread-safe)
func (c *Client) GetTelemetry() TelemetryData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.telemetry
}

// IsConnected returns true if connected to drone
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Consider disconnected if no heartbeat in 3 seconds
	if c.connected && time.Since(c.lastHeartbeat) > 3*time.Second {
		c.connected = false
		c.logger.Println("MAVLink: Connection timeout (no heartbeat)")
	}

	return c.connected
}

// IsArmed returns true if drone is armed
func (c *Client) IsArmed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.armed
}

// GetSystemID returns the drone's MAVLink system ID
func (c *Client) GetSystemID() uint8 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.systemID
}

// WaitForConnection waits for a heartbeat with timeout
func (c *Client) WaitForConnection(timeout time.Duration) error {
	c.logger.Printf("MAVLink: Waiting for heartbeat (timeout: %s)", timeout)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if c.IsConnected() {
			c.logger.Printf("MAVLink: Heartbeat received from system %d", c.GetSystemID())

			// Request data streams now that we're connected
			if err := c.requestDataStreams(); err != nil {
				c.logger.Printf("MAVLink: Warning - failed to request data streams: %v", err)
				// Non-fatal - continue anyway
			}

			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for heartbeat")
		}

		<-ticker.C
	}
}

// Arm sends arm command to the drone
func (c *Client) Arm() error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Println("MAVLink: Sending ARM command")

	return c.node.WriteMessageAll(&common.MessageCommandLong{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Command:         common.MAV_CMD_COMPONENT_ARM_DISARM,
		Param1:          1, // 1 = arm, 0 = disarm
	})
}

// Disarm sends disarm command to the drone
func (c *Client) Disarm() error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Println("MAVLink: Sending DISARM command")

	return c.node.WriteMessageAll(&common.MessageCommandLong{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Command:         common.MAV_CMD_COMPONENT_ARM_DISARM,
		Param1:          0, // 1 = arm, 0 = disarm
	})
}

// SetMode sets the flight mode using PX4's mode encoding
// The mode value is encoded in MAVLink's custom_mode field
func (c *Client) SetMode(px4Mode uint32) error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Printf("MAVLink: Setting PX4 mode to %d", px4Mode)

	// Send MAV_CMD_DO_SET_MODE command
	// Param1: MAV_MODE_FLAG_CUSTOM_MODE_ENABLED tells MAVLink to use custom_mode field
	// Param2: The PX4-specific mode value
	return c.node.WriteMessageAll(&common.MessageCommandLong{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Command:         common.MAV_CMD_DO_SET_MODE,
		Param1:          float32(common.MAV_MODE_FLAG_CUSTOM_MODE_ENABLED),
		Param2:          float32(px4Mode),
	})
}

// Takeoff sends takeoff command to the drone
func (c *Client) Takeoff(altitude float32) error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Printf("MAVLink: Sending TAKEOFF command (altitude: %.2fm)", altitude)

	return c.node.WriteMessageAll(&common.MessageCommandLong{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Command:         common.MAV_CMD_NAV_TAKEOFF,
		Param7:          altitude, // Target altitude
	})
}

// Land sends land command to the drone
func (c *Client) Land() error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Println("MAVLink: Sending LAND command")

	return c.node.WriteMessageAll(&common.MessageCommandLong{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Command:         common.MAV_CMD_NAV_LAND,
	})
}

// ReturnToLaunch sends RTL command to the drone
func (c *Client) ReturnToLaunch() error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Println("MAVLink: Sending RETURN_TO_LAUNCH command")

	return c.node.WriteMessageAll(&common.MessageCommandLong{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Command:         common.MAV_CMD_NAV_RETURN_TO_LAUNCH,
	})
}

// Close closes the MAVLink connection
func (c *Client) Close() error {
	c.logger.Println("MAVLink: Closing connection")

	// Stop ground station message sender
	close(c.stopHeartbeat)

	// Wait for goroutine to finish (with timeout)
	select {
	case <-c.heartbeatDone:
		c.logger.Println("MAVLink: Ground station message sender stopped")
	case <-time.After(2 * time.Second):
		c.logger.Println("MAVLink: Warning - ground station message sender stop timeout")
	}

	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	c.node.Close()
	return nil
}

// GetConnectionInfo returns connection information
func (c *Client) GetConnectionInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"port":           c.port,
		"baud_rate":      c.baudRate,
		"system_id":      c.systemID,
		"connected":      c.connected,
		"armed":          c.armed,
		"last_heartbeat": c.lastHeartbeat,
	}
}
