# Flightpath Server

Go backend for controlling drones through a unified, protocol-agnostic API.

## Key Features

✅ **Protocol-Agnostic Frontend** - Frontend only knows drone IDs  
✅ **Configuration-Driven** - All connection details in `drones.yaml`  
✅ **Multi-Protocol Support** - MAVLink, DJI, custom protocols  
✅ **Zero Frontend Changes** - Add drones by editing config file  
✅ **Production-Ready** - Proper separation of concerns  

## Quick Start
```bash
# 1. Clone repository
git clone https://github.com/flightpath-dev/flightpath-server
cd flightpath-server

# 2. Install dependencies
go mod tidy

# 3. Configure your drones (edit this file)
nano data/config/drones.yaml

# 4. Run server
go run cmd/server/main.go

# 5. Connect to drone (in another terminal)
./scripts/test.sh connect alpha
```

## Message Flow

Frontend:
1. Says "Connect to drone alpha"

Backend:
1. Looks up "alpha" in `drones.yaml`
2. Reads `mavlink` protocol, `/dev/cu.usbserial-D30JAXGS`, 57600 baud
3. Creates MAVLink client
4. Connects and returns success

Note: Frontend never knows about ports, protocols, or baud rates!

## Configuration

### Drone Registry

The `data/config/drones.yaml` file defines available drones. This file is committed to the repository and should be updated when adding new drones.

**`data/config/drones.yaml`**
```yaml
drones:
  - id: "alpha"
    name: "Alpha X500"
    description: "Primary test drone - Holybro X500 V2"
    protocol: "mavlink"
    connection:
      type: "serial"
      port: "/dev/cu.usbserial-D30JAXGS"
      baud_rate: 57600

  - id: "bravo"
    name: "Bravo Quadcopter"
    description: "Secondary test drone"
    protocol: "mavlink"
    connection:
      type: "serial"
      port: "/dev/ttyUSB1"
      baud_rate: 115200
```

### Data Directory Structure
```
data/
├── config/              # ✅ Version controlled - Configuration files
│   └── drones.yaml      # Drone registry
├── logs/                # ❌ Gitignored - Runtime logs
├── runtime/             # ❌ Gitignored - Runtime state
└── cache/               # ❌ Gitignored - Cached data
```

### Environment Variables

You can override configuration using environment variables:
```bash
# Server configuration
export FLIGHTPATH_HOST=0.0.0.0
export FLIGHTPATH_PORT=8080

# MAVLink defaults (used if not specified in drone config)
export FLIGHTPATH_MAVLINK_PORT=/dev/ttyUSB0
export FLIGHTPATH_MAVLINK_BAUD=57600

# Drone registry location
export FLIGHTPATH_DRONE_REGISTRY=./data/config/drones.yaml

# Logging
export FLIGHTPATH_LOG_LEVEL=info  # debug, info, warn, error
```

## Project Structure
```
flightpath-server/
├── cmd/
│   └── server/
│       └── main.go              # Server entry point
├── data/
│   └── config/
│       └── drones.yaml          # Drone configurations
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration types
│   │   ├── loader.go            # Environment variable loader
│   │   └── drones.go            # Drone registry loader
│   ├── mavlink/
│   │   └── client.go            # MAVLink protocol implementation
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware
│   │   ├── logging.go           # Request logging
│   │   └── recovery.go          # Panic recovery
│   ├── server/
│   │   ├── dependencies.go      # Shared dependencies
│   │   └── server.go            # HTTP server setup
│   └── services/
│       ├── connection.go        # Connection service (protocol routing)
│       ├── control.go           # Control service
│       ├── mission.go           # Mission service
│       └── telemetry.go         # Telemetry service
├── scripts/
│   └── test.sh                  # Helper script for testing
├── go.mod
└── go.sum
```

## API Services

### 1. ConnectionService

Manage drone connections by drone id.
```bash
# List all drones in registry
./scripts/test.sh list

# Connect to drone
./scripts/test.sh connect alpha

# Get connection status
./scripts/test.sh status alpha

# Disconnect
./scripts/test.sh disconnect alpha
```

### 2. ControlService

Send flight control commands.
```bash
# Arm drone (⚠️ REMOVE PROPELLERS FOR TESTING!)
./scripts/test.sh arm alpha

# Disarm drone
./scripts/test.sh disarm alpha

# Set flight mode
./scripts/test.sh mode alpha GUIDED

# Takeoff to 10 meters
./scripts/test.sh takeoff alpha 10

# Land
./scripts/test.sh land alpha

# Return home
./scripts/test.sh rtl alpha
```

### 3. TelemetryService

Stream real-time telemetry data from the drone.

**Features:**
- Real-time position (GPS coordinates, altitude)
- Velocity (3D velocity vector)
- Attitude (roll, pitch, yaw)
- Battery status (voltage, current, remaining %)
- System health (sensors, GPS)
- Flight mode
- GPS accuracy and satellite count
```bash
# Get telemetry snapshot (single point-in-time reading)
./scripts/test.sh snapshot alpha

# Monitor telemetry (continuous updates every 2 seconds)
./scripts/test.sh monitor alpha

# Stream telemetry requires HTTP/2 streaming client
# Best tested from frontend or tools like grpcurl
```

**Telemetry Data Available:**
- **Position**: Latitude, longitude, altitude (MSL)
- **Velocity**: North, east, down components (m/s)
- **Attitude**: Roll, pitch, yaw (radians)
- **Battery**: Voltage (V), current (A), remaining (%)
- **Health**: Sensor status, GPS status
- **Navigation**: Heading (°), ground speed (m/s), vertical speed (m/s)
- **GPS**: Accuracy (m), satellite count
- **Status**: Armed state, flight mode

### 4. MissionService 🚧 Skeleton Implementation

Autonomous mission planning and execution (stubs for future implementation).

## Flight Modes for API Control

Flightpath is designed for API-controlled flight **without RC transmitter**. Understanding flight modes is critical for safe operation.

### GUIDED Mode (Recommended for API Control)

**Use for:** Dynamic position commands from the API

**What is GUIDED?**
- Holds GPS position automatically (no drift)
- **Accepts position/velocity commands from API** ✅
- Responds immediately to `GoToPosition` commands
- When not commanded, hovers safely in place
- Think: "Position hold + API control enabled"

**Stay in GUIDED for entire flight:**
```bash
# 1. Connect and arm
./scripts/test.sh connect alpha
./scripts/test.sh arm alpha

# 2. Set GUIDED mode
./scripts/test.sh mode alpha GUIDED

# 3. Takeoff
./scripts/test.sh takeoff alpha 20

# 4. Send position commands (no mode changes needed!)
# Drone responds immediately and holds position when idle

# 5. Land
./scripts/test.sh land alpha

# 6. Disarm
./scripts/test.sh disarm alpha
```

**Why GUIDED?**
- ✅ Can send position commands at any time
- ✅ Holds position safely when not commanded
- ✅ No mode switching between commands
- ✅ Most responsive to API commands

### POSITION_HOLD Mode (Safety Lockdown)

**Use for:** Preventing API commands (safety feature)

**What is POSITION_HOLD?**
- Holds GPS position automatically (no drift)
- **Rejects position/velocity commands from API** ❌
- Only responds to RC stick inputs (if connected)
- Think: "Hold position and ignore external commands"
```bash
# Switch to POSITION_HOLD to "freeze" drone
./scripts/test.sh mode alpha POSITION_HOLD
```

**When to use:**
- 🔒 Lock drone in place (prevent buggy API commands)
- 🛑 Pause API control temporarily
- ⏸️ Debugging/development pause

**Key Difference:**
```
Scenario: API sends GoToPosition command

GUIDED mode:
  → ✅ "Roger, flying to new position"

POSITION_HOLD mode:
  → ❌ "Ignoring command, holding current position"
```

### AUTO Mode (Mission Control) ⭐ Safest for Autonomous

**Use for:** Pre-programmed waypoint missions

**What is AUTO?**
- Follows uploaded waypoint mission
- Fully autonomous (takeoff → waypoints → land)
- Safest for predictable, pre-planned flights
- No real-time API commands needed

**Mission workflow:**
```bash
# 1. Create mission file with waypoints
cat > mission.json << 'EOF'
{
  "mission": {
    "id": "survey-001",
    "name": "Area Survey",
    "waypoints": [
      {
        "sequence": 0,
        "action": "ACTION_TAKEOFF",
        "position": {"latitude": 37.7749, "longitude": -122.4194, "altitude": 20}
      },
      {
        "sequence": 1,
        "action": "ACTION_WAYPOINT",
        "position": {"latitude": 37.7750, "longitude": -122.4195, "altitude": 30}
      },
      {
        "sequence": 2,
        "action": "ACTION_LAND",
        "position": {"latitude": 37.7749, "longitude": -122.4194, "altitude": 0}
      }
    ]
  }
}
EOF

# 2. Connect and upload mission
./scripts/test.sh connect alpha
curl -X POST --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d @mission.json \
  http://localhost:8080/drone.v1.MissionService/UploadMission

# 3. Arm and start mission
./scripts/test.sh arm alpha
curl -X POST --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.MissionService/StartMission

# Mission runs automatically: takeoff → waypoints → land
# No position commands needed!
```

**Why AUTO for missions?**
- ✅ Pre-planned safe path (reviewed before flight)
- ✅ Includes takeoff and landing waypoints
- ✅ Most predictable behavior
- ✅ Best for unattended operations
- ✅ Automatic failsafes (geofence, battery RTL)

### RTL Mode (Return to Launch)

**Use for:** Emergency return home

**What is RTL?**
- Autonomous return to launch position
- Climbs to safe altitude, flies home, lands
- Triggered automatically on low battery or geofence breach
- Can be commanded via API
```bash
# Emergency return home
./scripts/test.sh rtl alpha
```

### Other Modes

Additional modes available for specific use cases:

**STABILIZED** - Attitude stabilization only (manual control)
```bash
./scripts/test.sh mode alpha STABILIZED
```

**ALTITUDE_HOLD** - Holds altitude, manual horizontal control
```bash
./scripts/test.sh mode alpha ALTITUDE_HOLD
```

**MANUAL** - Full manual control (no stabilization)
```bash
./scripts/test.sh mode alpha MANUAL
```

**LOITER** - Circle around current position
```bash
./scripts/test.sh mode alpha LOITER
```

### Mode Comparison

| Mode | Accepts API Commands | Best For | Safety |
|------|---------------------|----------|--------|
| **GUIDED** | ✅ Yes | Dynamic API control | ⭐⭐⭐⭐ |
| **POSITION_HOLD** | ❌ No | Safety lockdown | ⭐⭐⭐⭐⭐ |
| **AUTO (Mission)** | ❌ No (follows mission) | Pre-planned autonomous | ⭐⭐⭐⭐⭐ |
| **RTL** | ❌ No (autonomous return) | Emergency return | ⭐⭐⭐⭐⭐ |
| **ALTITUDE_HOLD** | ❌ No | Manual flight with altitude hold | ⭐⭐⭐ |
| **STABILIZED** | ❌ No | Manual flight with stabilization | ⭐⭐ |
| **MANUAL** | ❌ No | Full manual control | ⭐ |
| **LOITER** | ❌ No | Circle position | ⭐⭐⭐⭐ |

### PX4 Mode Mapping Reference

Flightpath uses generic flight mode names that map to PX4-specific modes:

| Flightpath Mode | PX4 Main Mode | PX4 Sub Mode | PX4 Mode Value | Description |
|-----------------|---------------|--------------|----------------|-------------|
| `MANUAL` | `MANUAL (1)` | - | `1` | Full manual control |
| `STABILIZED` | `STABILIZED (7)` | - | `7` | Attitude stabilization |
| `ALTITUDE_HOLD` | `ALTCTL (2)` | - | `2` | Altitude control |
| `POSITION_HOLD` | `POSCTL (3)` | - | `3` | GPS position hold |
| `GUIDED` | `OFFBOARD (6)` | - | `6` | External position control (API) |
| `AUTO` | `AUTO (4)` | `MISSION (4)` | `262148` | Follow mission waypoints |
| `RETURN_HOME` | `AUTO (4)` | `RTL (5)` | `327684` | Return to launch |
| `LAND` | `AUTO (4)` | `LAND (6)` | `393220` | Land at position |
| `TAKEOFF` | `AUTO (4)` | `TAKEOFF (2)` | `131076` | Takeoff |
| `LOITER` | `AUTO (4)` | `LOITER (3)` | `196612` | Circle position |

**PX4 Mode Encoding:**
- Simple modes: Use main mode value directly
- AUTO sub-modes: `main_mode | (sub_mode << 16)`
  - Example: RTL = `4 | (5 << 16)` = `327684`

### Recommended Approach

**For API Control:**
1. Stay in **GUIDED mode** for entire flight
2. Drone holds position when not commanded (safe)
3. Responds immediately to position commands
4. Switch to **POSITION_HOLD** only to "freeze" drone

**For Autonomous Operations:**
1. Use **AUTO mode** with pre-programmed missions
2. Upload mission including takeoff and landing
3. Start mission and monitor progress
4. Most predictable and safest for unattended flight

## Flight Safety

### Pre-Flight Checklist (No RC)

Before flying with API only:

1. ⚠️ **Remove propellers** for initial testing
2. ✅ Test mission in simulator first (if using AUTO mode)
3. ✅ Verify GPS lock (satellite count ≥ 8)
4. ✅ Check battery level (> 50% recommended)
5. ✅ Confirm geofence is set correctly
6. ✅ Set home position (for RTL)
7. ✅ Verify clear flight area
8. ✅ Test Return Home procedure
9. ✅ Monitor telemetry during flight

### Safe Command Sequence (API Control)
```bash
# Complete flight sequence with GUIDED mode

# 1. Connect to drone
./scripts/test.sh connect alpha

# 2. Check telemetry before flight
./scripts/test.sh snapshot alpha

# 3. Arm drone
./scripts/test.sh arm alpha

# 4. Set GUIDED mode (accepts API commands)
./scripts/test.sh mode alpha GUIDED

# 5. Takeoff
./scripts/test.sh takeoff alpha 20

# 6. Monitor telemetry during flight
./scripts/test.sh monitor alpha

# 7. Send position commands as needed
# (GoToPosition not yet implemented)

# 8. Land
./scripts/test.sh land alpha

# 9. Disarm
./scripts/test.sh disarm alpha

# 10. Disconnect
./scripts/test.sh disconnect alpha
```

### Emergency Procedures (No RC)

**If something goes wrong:**

1. **Return Home** (Most Common)
```bash
   ./scripts/test.sh rtl alpha
```

2. **Hold Position** (Stop and hover - switch to POSITION_HOLD)
```bash
   ./scripts/test.sh mode alpha POSITION_HOLD
```

3. **Emergency Land** (Land immediately at current location)
```bash
   ./scripts/test.sh land alpha
```

**⚠️ DO NOT use EmergencyStop - it cuts motors and drone will fall!**

### Automatic Failsafes

PX4/ArduPilot automatically handles:

- **Low Battery** → Triggers RTL automatically
- **GPS Lost** → Enters ALTITUDE_HOLD and hovers
- **Geofence Breach** → Triggers RTL or LAND
- **Datalink Loss** → Configured failsafe action (default: RTL)

### API Connection Loss

If API/network connection is lost:

- **GUIDED mode** → Continues hovering at last position (safe)
- **AUTO mode** → Continues mission as programmed
- If connection lost > timeout → Triggers failsafe (RTL)

## Adding a New Drone

### Step 1: Edit Configuration

Add your drone to `data/config/drones.yaml`:
```yaml
drones:
  # ... existing drones ...
  
  - id: "charlie"
    name: "Charlie Custom"
    description: "Custom built quadcopter"
    protocol: "mavlink"
    connection:
      type: "serial"
      port: "/dev/ttyUSB2"
      baud_rate: 115200
```

### Step 2: Restart Server
```bash
go run cmd/server/main.go
```

### Step 3: Connect
```bash
./scripts/test.sh connect charlie
```

**No code changes needed!**

## Supported Protocols

- ✅ **MAVLink** (PX4, ArduPilot)
  - Serial connection (USB, UART)
  - UDP connection (for simulators)
  - Full flight mode control
  - Arm/Disarm, Takeoff/Land, RTL
  - Real-time telemetry streaming
- 🔜 **DJI SDK** - Planned
- 🔜 **Custom** - Extensible architecture

## Frontend Example
```typescript
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ConnectionService } from "@flightpath-dev/flightpath-proto/gen/ts/drone/v1/connection_connect";

// Create transport
const transport = createConnectTransport({
  baseUrl: "http://localhost:8080",
});

// Create client
const client = createClient(ConnectionService, transport);

// Connect - that's all the frontend needs to know!
const response = await client.connect({
  droneId: "alpha"
});

console.log(response.message); 
// "Connected to Alpha X500 (System ID: 1)"
```

## Testing

### Complete Test Flow
```bash
# 1. Start server
go run cmd/server/main.go

# 2. In another terminal - test connection
./scripts/test.sh list                # List all drones
./scripts/test.sh connect alpha       # Connect to alpha
./scripts/test.sh status alpha        # Check status

# 3. Test telemetry
./scripts/test.sh snapshot alpha      # Get single telemetry reading
./scripts/test.sh monitor alpha       # Monitor telemetry (Ctrl+C to stop)

# 4. Test flight modes (propellers off!)
./scripts/test.sh arm alpha                  # Arm
./scripts/test.sh mode alpha GUIDED          # Set GUIDED mode
./scripts/test.sh mode alpha POSITION_HOLD   # Set POSITION_HOLD
./scripts/test.sh mode alpha AUTO            # Set AUTO mode
./scripts/test.sh disarm alpha               # Disarm

# 5. Full flight test (after successful mode tests)
./scripts/test.sh arm alpha           # Arm
./scripts/test.sh mode alpha GUIDED   # Set GUIDED mode
./scripts/test.sh takeoff alpha 10    # Takeoff to 10m
./scripts/test.sh land alpha          # Land
./scripts/test.sh disarm alpha        # Disarm

# 6. Emergency procedures
./scripts/test.sh rtl alpha           # Return home
./scripts/test.sh mode alpha POSITION_HOLD  # Hold position

# 7. Cleanup
./scripts/test.sh disconnect alpha    # Disconnect
```

### Available Test Commands
```bash
./scripts/test.sh list                          # List drones
./scripts/test.sh connect <drone_id>            # Connect to drone
./scripts/test.sh disconnect <drone_id>         # Disconnect
./scripts/test.sh status <drone_id>             # Get status
./scripts/test.sh snapshot <drone_id>           # Get telemetry snapshot
./scripts/test.sh monitor <drone_id>            # Monitor telemetry (live)
./scripts/test.sh arm <drone_id>                # Arm
./scripts/test.sh disarm <drone_id>             # Disarm
./scripts/test.sh mode <drone_id> <MODE>        # Set flight mode
./scripts/test.sh takeoff <drone_id> <alt>      # Takeoff
./scripts/test.sh land <drone_id>               # Land
./scripts/test.sh rtl <drone_id>                # Return home
```

**Available modes:** `MANUAL`, `STABILIZED`, `ALTITUDE_HOLD`, `POSITION_HOLD`, `GUIDED`, `AUTO`, `RETURN_HOME`, `LAND`, `TAKEOFF`, `LOITER`

## Development

### Install Dependencies
```bash
# Install Go dependencies
go mod tidy

# Install buf (for proto generation)
brew install bufbuild/buf/buf  # macOS
# or
go install github.com/bufbuild/buf/cmd/buf@latest
```

### Update Proto Definitions

Proto definitions are in a separate repository: [`flightpath-proto`](https://github.com/flightpath-dev/flightpath-proto)

To update to a new proto version:
```bash
# Update proto dependency
go get -u github.com/flightpath-dev/flightpath-proto@latest

# Update go.mod
go mod tidy

# Restart server
go run cmd/server/main.go
```

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config
```

## Troubleshooting

### "Drone not found in registry"

Check that your drone ID exists in `data/config/drones.yaml`:
```bash
# List available drones
./scripts/test.sh list
```

### "Failed to create MAVLink connection"

1. Check serial port exists:
```bash
   # Linux
   ls /dev/ttyUSB*
   
   # macOS
   ls /dev/tty.usbserial-*
```

2. Check permissions:
```bash
   # Linux - add user to dialout group
   sudo usermod -a -G dialout $USER
   
   # macOS - no special permissions needed
```

3. Verify baud rate matches your drone's configuration (usually 57600 or 115200)

### "Connection timeout"

1. Check drone is powered on
2. Verify serial cable is connected
3. Confirm baud rate in `drones.yaml` matches drone settings
4. Test with QGroundControl first to verify hardware connection

### "Mode change failed" or "Command denied"

1. Check drone is armed (some modes require armed state)
2. Verify GPS lock for GPS-dependent modes (GUIDED, POSITION_HOLD, AUTO, RTL)
3. Check pre-arm checks passed
4. Review drone logs for specific error messages

### "No telemetry data" or "Zero values"

1. Ensure drone is fully powered on and booted
2. Wait for GPS lock (satellite count ≥ 6)
3. Verify MAVLink messages are being received (check server logs)
4. Some telemetry requires GPS lock before reporting valid data

### Port already in use
```bash
# Check what's using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
export FLIGHTPATH_PORT=8081
go run cmd/server/main.go
```

## Roadmap

- **Iteration 1** ✅ - Connection and basic control (MAVLink)
- **Iteration 2** ✅ - Protocol-agnostic architecture  
- **Iteration 3** ✅ - Flight mode control (GUIDED, POSITION_HOLD, AUTO, RTL)
- **Iteration 4** ✅ - Real-time telemetry streaming
- **Iteration 5** 📋 - Mission planning and waypoints
- **Iteration 6** 📋 - Position commands (GoToPosition)
- **Iteration 7** 📋 - React frontend
- **Iteration 8** 📋 - Authentication

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

For proto changes, see the [`flightpath-proto`](https://github.com/flightpath-dev/flightpath-proto) repository.

## Support

For issues or questions:
- Open an issue on GitHub
- Check existing documentation
- Review the proto definitions at [`flightpath-proto`](https://github.com/flightpath-dev/flightpath-proto)