# Flightpath Server - Protocol-Agnostic Drone Control

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
# Using curl (simpler, no schema required):
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{"drone_id": "alpha"}' \
  http://localhost:8080/drone.v1.ConnectionService/Connect

# Alternative: Using buf curl (requires schema):
# git clone https://github.com/flightpath-dev/flightpath-proto
# buf curl --http2-prior-knowledge \
#   --protocol connect \
#   --schema <path-to-flightpath-proto> \
#   --data '{"drone_id": "alpha"}' \
#   http://localhost:8080/drone.v1.ConnectionService/Connect
```

## Architecture

Frontend says: **"Connect to drone alpha"**

Backend does:
1. Looks up "alpha" in `drones.yaml`
2. Reads: MAVLink protocol, `/dev/ttyUSB0`, 57600 baud
3. Creates MAVLink client
4. Connects and returns success

**Frontend never knows about ports, protocols, or baud rates!**

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
      port: "/dev/ttyUSB0"
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
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration types
│   │   ├── loader.go            # Environment variable loader
│   │   └── drones.go            # Drone registry loader
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware
│   │   ├── logging.go           # Request logging
│   │   └── recovery.go          # Panic recovery
│   ├── mavlink/
│   │   └── client.go            # MAVLink protocol implementation
│   ├── server/
│   │   ├── dependencies.go      # Shared dependencies
│   │   └── server.go            # HTTP server setup
│   └── services/
│       ├── connection.go        # Connection service (protocol routing)
│       ├── control.go           # Control service
│       ├── telemetry.go         # Telemetry service (skeleton)
│       └── mission.go           # Mission service (skeleton)
├── data/
│   └── config/
│       └── drones.yaml          # Drone configurations
├── scripts/
│   └── test.sh                  # Helper script for testing
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

## API Services

**Note:** All examples use `curl` for simplicity. You can also use `buf curl` with the `--schema` flag if you prefer.

### 1. ConnectionService ✅ Fully Implemented

Manage drone connections by ID only.
```bash
# Connect to drone
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{"drone_id": "alpha"}' \
  http://localhost:8080/drone.v1.ConnectionService/Connect

# Get connection status
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ConnectionService/GetStatus

# List available drones
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ConnectionService/ListDrones

# Disconnect
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ConnectionService/Disconnect
```

### 2. ControlService ✅ Fully Implemented

Send flight control commands.
```bash
# Arm drone (⚠️ REMOVE PROPELLERS FOR TESTING!)
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ControlService/Arm

# Disarm drone
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ControlService/Disarm

# Takeoff to 10 meters
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{"altitude": 10}' \
  http://localhost:8080/drone.v1.ControlService/Takeoff

# Land
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ControlService/Land

# Return home
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ControlService/ReturnHome
```

### 3. TelemetryService 🚧 Skeleton Implementation

Stream real-time telemetry data (basic implementation).
```bash
# Get telemetry snapshot
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.TelemetryService/GetSnapshot
```

### 4. MissionService 🚧 Skeleton Implementation

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
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{"drone_id": "charlie"}' \
  http://localhost:8080/drone.v1.ConnectionService/Connect
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
./scripts/test.sh list                # List drones
./scripts/test.sh connect alpha       # Connect to alpha
./scripts/test.sh status              # Check status
./scripts/test.sh arm                 # Arm
./scripts/test.sh takeoff 10          # Takeoff to 10m
./scripts/test.sh land                # Land
./scripts/test.sh rtl                 # Return home
./scripts/test.sh disarm              # Disarm
./scripts/test.sh disconnect          # Disconnect
```

**Note:** The script uses the `--http2-prior-knowledge` flag. This is required for development because we're using HTTP instead of HTTPS. In production with HTTPS, this flag is not needed.

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

### "failed to find service named 'drone.v1.ConnectionService' in schema"

This error occurs when using `buf curl` without providing the schema. The server doesn't have reflection enabled.

**Solution:** Use regular `curl` instead (simpler, no schema required):
```bash
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{"drone_id": "alpha"}' \
  http://localhost:8080/drone.v1.ConnectionService/Connect
```

If you prefer `buf curl`, you'll need to provide the schema:
```bash
git clone https://github.com/flightpath-dev/flightpath-proto
buf curl --http2-prior-knowledge \
  --protocol connect \
  --schema <path-to-flightpath-proto> \
  --data '{"drone_id": "alpha"}' \
  http://localhost:8080/drone.v1.ConnectionService/Connect
```

### "Drone not found in registry"

Check that your drone ID exists in `data/config/drones.yaml`:
```bash
# List available drones
curl -X POST \
  --http2-prior-knowledge \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ConnectionService/ListDrones
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