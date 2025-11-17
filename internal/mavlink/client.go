package mavlink

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
)

// PX4 Custom Main Modes
const (
	PX4_CUSTOM_MAIN_MODE_MANUAL     = 1
	PX4_CUSTOM_MAIN_MODE_ALTCTL     = 2
	PX4_CUSTOM_MAIN_MODE_POSCTL     = 3
	PX4_CUSTOM_MAIN_MODE_AUTO       = 4
	PX4_CUSTOM_MAIN_MODE_ACRO       = 5
	PX4_CUSTOM_MAIN_MODE_OFFBOARD   = 6
	PX4_CUSTOM_MAIN_MODE_STABILIZED = 7
	PX4_CUSTOM_MAIN_MODE_RATTITUDE  = 8
)

// PX4 Custom Sub Modes (for AUTO mode)
const (
	PX4_CUSTOM_SUB_MODE_AUTO_READY    = 1
	PX4_CUSTOM_SUB_MODE_AUTO_TAKEOFF  = 2
	PX4_CUSTOM_SUB_MODE_AUTO_LOITER   = 3
	PX4_CUSTOM_SUB_MODE_AUTO_MISSION  = 4
	PX4_CUSTOM_SUB_MODE_AUTO_RTL      = 5
	PX4_CUSTOM_SUB_MODE_AUTO_LAND     = 6
	PX4_CUSTOM_SUB_MODE_AUTO_FOLLOW   = 8
	PX4_CUSTOM_SUB_MODE_AUTO_PRECLAND = 9
)

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
	}

	// Start listening for messages
	go client.listen()

	return client, nil
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

		// Add more message handlers as needed
		// case *common.MessageGlobalPositionInt:
		//     c.handleGlobalPosition(m)
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

// SetMode sets the flight mode using PX4 custom mode
func (c *Client) SetMode(customMode uint32) error {
	c.mu.RLock()
	systemID := c.systemID
	c.mu.RUnlock()

	if !c.IsConnected() {
		return fmt.Errorf("not connected to drone")
	}

	c.logger.Printf("MAVLink: Setting mode to custom mode %d", customMode)

	// Send MAV_CMD_DO_SET_MODE command
	// Param1: Mode (MAV_MODE_FLAG_CUSTOM_MODE_ENABLED = 1)
	// Param2: Custom mode
	return c.node.WriteMessageAll(&common.MessageCommandLong{
		TargetSystem:    systemID,
		TargetComponent: 1,
		Command:         common.MAV_CMD_DO_SET_MODE,
		Param1:          float32(common.MAV_MODE_FLAG_CUSTOM_MODE_ENABLED),
		Param2:          float32(customMode),
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
