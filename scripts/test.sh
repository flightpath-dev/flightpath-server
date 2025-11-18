#!/bin/bash

# Helper script for testing Flightpath server

# Using `curl` (simpler, no schema required)
#
# Alternative: Using `buf curl` (requires schema):
# git clone https://github.com/flightpath-dev/flightpath-proto
# buf curl --http2-prior-knowledge \
#   --protocol connect \
#   --schema <path-to-flightpath-proto> \
#   --data '{"drone_id": "alpha"}' \
#   http://localhost:8080/drone.v1.ConnectionService/Connect
#
# **Note:** This script uses the `--http2-prior-knowledge` flag. This is required for development because we're using HTTP instead of HTTPS.
# In production with HTTPS, this flag is not needed.


URL="http://localhost:8080"

case "$1" in
  list)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ConnectionService/ListDrones
    ;;
  connect)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ConnectionService/Connect
    ;;
  disconnect)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ConnectionService/Disconnect
    ;;
  status)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ConnectionService/GetStatus
    ;;
  snapshot)
    echo "📊 Telemetry Snapshot for $2:"
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.TelemetryService/GetSnapshot | jq '.'
    ;;
  monitor)
    echo "📡 Monitoring telemetry for $2 (Ctrl+C to stop)..."
    echo "Press Ctrl+C to stop monitoring"
    echo ""
    while true; do
      clear
      echo "=== Telemetry Monitor for $2 ==="
      echo "$(date)"
      echo ""
      curl -s -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.TelemetryService/GetSnapshot | jq '{
        armed: .armed,
        mode: .mode,
        position: {
          lat: .position.latitude,
          lon: .position.longitude,
          alt: (.position.altitude | tonumber | . * 100 | round / 100)
        },
        battery: {
          voltage: (.battery.voltage | tonumber | . * 100 | round / 100),
          remaining: .battery.remaining
        },
        gps: {
          satellites: .satellite_count,
          accuracy: (.gps_accuracy | tonumber | . * 100 | round / 100)
        },
        velocity: {
          ground_speed: (.ground_speed | tonumber | . * 100 | round / 100),
          vertical_speed: (.vertical_speed | tonumber | . * 100 | round / 100)
        }
      }'
      sleep 2
    done
    ;;
  arm)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ControlService/Arm
    ;;
  disarm)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ControlService/Disarm
    ;;
  mode)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\", \"mode\": \"FLIGHT_MODE_$3\"}" $URL/drone.v1.ControlService/SetFlightMode
    ;;
  takeoff)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\", \"altitude\": $3}" $URL/drone.v1.ControlService/Takeoff
    ;;
  land)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ControlService/Land
    ;;
  rtl)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ControlService/ReturnHome
    ;;
  *)
    echo "Usage: $0 {list|connect <drone_id>|status <drone_id>|snapshot <drone_id>|monitor <drone_id>|arm <drone_id>|disarm <drone_id>|mode <drone_id> <mode>|takeoff <drone_id> <altitude>|land <drone_id>|rtl <drone_id>|disconnect <drone_id>}"
    echo ""
    echo "Commands:"
    echo "  list                           - List all drones in registry"
    echo "  connect <drone_id>             - Connect to drone"
    echo "  disconnect <drone_id>          - Disconnect from drone"
    echo "  status <drone_id>              - Get connection status"
    echo "  snapshot <drone_id>            - Get single telemetry reading"
    echo "  monitor <drone_id>             - Monitor telemetry continuously"
    echo "  arm <drone_id>                 - Arm motors"
    echo "  disarm <drone_id>              - Disarm motors"
    echo "  mode <drone_id> <MODE>         - Set flight mode"
    echo "  takeoff <drone_id> <altitude>  - Takeoff to altitude (meters)"
    echo "  land <drone_id>                - Land at current position"
    echo "  rtl <drone_id>                 - Return to launch"
    echo ""
    echo "Available Modes:"
    echo "  MANUAL, STABILIZED, ALTITUDE_HOLD, POSITION_HOLD, GUIDED,"
    echo "  AUTO, RETURN_HOME, LAND, TAKEOFF, LOITER"
    echo ""
    echo "Examples:"
    echo "  $0 connect alpha"
    echo "  $0 snapshot alpha"
    echo "  $0 monitor alpha"
    echo "  $0 mode alpha GUIDED"
    echo "  $0 takeoff alpha 10"
    exit 1
    ;;
esac