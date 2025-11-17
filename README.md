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

### 1. ConnectionService ✅ Fully Implemented

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

### 2. ControlService ✅ Fully Implemented

Send flight control commands.
```bash
# Arm drone (⚠️ REMOVE PROPELLERS FOR TESTING!)
./scripts/test.sh arm alpha

# Disarm drone
./scripts/test.sh disarm alpha

# Takeoff to 10 meters
./scripts/test.sh takeoff alpha 10

# Land
./scripts/test.sh land alpha

# Return home
./scripts/test.sh rtl alpha
```

### 3. TelemetryService 🚧 Implementation

Stream real-time telemetry data (basic implementation).
```bash
# Get telemetry snapshot
./scripts/test.sh snapshot alpha
```

### 4. MissionService 🚧 Implementation

Autonomous mission planning and execution (stubs for future implementation).

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

- ✅ **MAVLink** (PX4, ArduPilot) - Fully implemented
  - Serial connection (USB, UART)
  - UDP connection (for simulators)
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

# 2. In another terminal
```bash
./scripts/test.sh list                # List all drones in registry
./scripts/test.sh connect alpha       # Connect to alpha
./scripts/test.sh status alpha        # Check status
./scripts/test.sh snapshot alpha      # Get telemetry snapshot
./scripts/test.sh arm alpha           # Arm
./scripts/test.sh takeoff alpha 10    # Takeoff to 10m
./scripts/test.sh land alpha          # Land
./scripts/test.sh rtl alpha           # Return home
./scripts/test.sh disarm alpha        # Disarm
./scripts/test.sh disconnect alpha    # Disconnect
```

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
go get github.com/flightpath-dev/flightpath-proto@v0.2.0

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
- **Iteration 3** 📋 - Real-time telemetry streaming
- **Iteration 4** 📋 - Mission planning and waypoints
- **Iteration 5** 📋 - React frontend
- **Iteration 6** 📋 - Authentication

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