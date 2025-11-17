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
  arm)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ControlService/Arm
    ;;
  disarm)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ControlService/Disarm
    ;;
  mode)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\", \"mode\": FLIGHT_MODE_$3}" $URL/drone.v1.ControlService/SetFlightMode
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
  snapshot)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.TelemetryService/GetSnapshot
    ;;
  *)
    echo "Usage: $0 {list|connect <drone_id>|status <drone_id>|arm <drone_id>|disarm <drone_id>|mode <drone_id> <mode>|takeoff <drone_id> <altitude>|land <drone_id>|rtl <drone_id>|disconnect <drone_id>|snapshot <drone_id>}"
    echo ""
    echo "Modes: MANUAL, STABILIZED, ALTITUDE_HOLD, POSITION_HOLD, GUIDED, AUTO, RETURN_HOME, LAND, TAKEOFF, LOITER"
    exit 1
    ;;
esac
