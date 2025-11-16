# Flightpath Server

Go backend for the Flightpath drone control platform.

## Prerequisites

- **Go 1.21 or later** (see `go.mod` for the exact version)
  ```bash
  # Check version
  go version
  ```

## Quick Start

```bash
# Clone the repository
git clone https://github.com/flightpath-dev/flightpath-server
cd flightpath-server

# Install dependencies
go mod tidy

# Run server
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

## Configuration

Configure via environment variables:

```bash
export FLIGHTPATH_PORT=8080
export FLIGHTPATH_HOST=0.0.0.0
export FLIGHTPATH_LOG_LEVEL=info
export FLIGHTPATH_MAVLINK_PORT=/dev/ttyUSB0
export FLIGHTPATH_MAVLINK_BAUD=57600
```

Or use defaults (defined in `internal/config/config.go`).

## Testing

Test the server endpoints with `curl`:

```bash
# Test connection endpoint
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"port": "/dev/ttyUSB0", "baud_rate": 57600}' \
  http://localhost:8080/drone.v1.ConnectionService/Connect

# Test status endpoint
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ConnectionService/GetStatus

# Test arm endpoint
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8080/drone.v1.ControlService/Arm
```

**Note**: You can also use `buf curl` if you have Buf CLI installed, but it's not required.

## Development

```bash
# Run with debug logging
FLIGHTPATH_LOG_LEVEL=debug go run cmd/server/main.go

# Build binary
go build -o flightpath-server cmd/server/main.go

# Run binary
./flightpath-server

# Build with optimizations
go build -ldflags="-s -w" -o flightpath-server cmd/server/main.go
```

### Development Workflow

```bash
# Terminal 1: Run server with hot reload (optional)
# Install air (hot reload tool)
go install github.com/cosmtrek/air@latest

# Run with air
air

# Terminal 2: Test endpoints (see Testing section above)
```

## Building for Production

```bash
# Build binary
go build -o flightpath-server cmd/server/main.go

# Run binary
./flightpath-server

# Or build with optimizations
go build -ldflags="-s -w" -o flightpath-server cmd/server/main.go
```

## Updating Dependencies

When proto definitions or other dependencies are updated:

```bash
# Update all dependencies to latest compatible versions
go get -u ./...

# Or update specific dependency (e.g., proto module)
go get -u github.com/flightpath-dev/flightpath-proto@latest

# Tidy up Go modules
go mod tidy
```

## Architecture

```
cmd/
  └── server/        - Entry point
internal/
  ├── config/        - Configuration management
  ├── middleware/    - HTTP middleware (CORS, logging, recovery)
  ├── server/        - Server setup and dependencies
  └── services/      - Connect service implementations
```

The proto definitions are managed as a Go module dependency (`github.com/flightpath-dev/flightpath-proto`). No local code generation is required.

## Troubleshooting

### "cannot find package" errors

```bash
# Download all dependencies
go mod download

# Tidy up Go modules
go mod tidy
```

### Port already in use

Change the port:
```bash
FLIGHTPATH_PORT=9090 go run cmd/server/main.go
```

### Permission denied on /dev/ttyUSB0

```bash
# Add your user to dialout group (Linux)
sudo usermod -a -G dialout $USER

# Or temporarily change permissions
sudo chmod 666 /dev/ttyUSB0
```

## Need Help?

- Check GitHub issues: https://github.com/flightpath-dev/flightpath-server/issues
- Review proto definitions: https://github.com/flightpath-dev/flightpath-proto
- Connect protocol docs: https://connectrpc.com/docs/go/getting-started
