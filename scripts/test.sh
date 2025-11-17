#!/bin/bash

# Helper script for testing Flightpath server

URL="http://localhost:8080"

case "$1" in
  list)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ConnectionService/ListDrones
    ;;
  connect)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"drone_id\": \"$2\"}" $URL/drone.v1.ConnectionService/Connect
    ;;
  status)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ConnectionService/GetStatus
    ;;
  arm)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ControlService/Arm
    ;;
  disarm)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ControlService/Disarm
    ;;
  takeoff)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d "{\"altitude\": $2}" $URL/drone.v1.ControlService/Takeoff
    ;;
  land)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ControlService/Land
    ;;
  rtl)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ControlService/ReturnHome
    ;;
  disconnect)
    curl -X POST --http2-prior-knowledge -H "Content-Type: application/json" -d '{}' $URL/drone.v1.ConnectionService/Disconnect
    ;;
  *)
    echo "Usage: $0 {list|connect <drone_id>|status|arm|disarm|takeoff <altitude>|land|rtl|disconnect}"
    exit 1
    ;;
esac
