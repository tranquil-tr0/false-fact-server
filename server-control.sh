#!/bin/bash

# script for false fact backend/api/server control
# Usage: ./server-control.sh [start|stop|restart|status]

SERVER_NAME="server"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PID_FILE="$SCRIPT_DIR/.${SERVER_NAME}.pid"
LOG_FILE="$SCRIPT_DIR/${SERVER_NAME}.log"

# Load PORT from .env file
if [ -f "$SCRIPT_DIR/.env" ]; then
    export $(grep -E '^PORT=' "$SCRIPT_DIR/.env" | xargs)
fi

start_server() {
    if [ -f "$PID_FILE" ]; then
        echo "Server is already running (PID: $(cat $PID_FILE))"
        return 1
    fi

    # Check if port is available
    if netstat -tuln | grep -q ":$PORT "; then
        echo "Error: Port $PORT is already in use!"
        echo "Cannot start server: port defined in .env is busy."
        return 1
    fi

    echo "Starting $SERVER_NAME on port $PORT..."
    nohup ./$SERVER_NAME > "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    echo "Server started with PID: $!"
    echo "Logs available in $LOG_FILE"
    echo "Using port: $PORT"
}

stop_server() {
    if [ ! -f "$PID_FILE" ]; then
        echo "Server is not running"
        return 1
    fi
    
    PID=$(cat "$PID_FILE")
    echo "Stopping server (PID: $PID)..."
    kill $PID
    rm -f "$PID_FILE"
    echo "Server stopped"
}

restart_server() {
    stop_server
    sleep 2
    start_server
}

status_server() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p $PID > /dev/null; then
            echo "Server is running (PID: $PID)"
            echo "Using port: $PORT"
            echo "Log file: $LOG_FILE"
        else
            echo "PID file exists but process is not running"
            rm -f "$PID_FILE"
        fi
    else
        echo "Server is not running"
    fi
}

case "$1" in
    start)
        start_server
        ;;
    stop)
        stop_server
        ;;
    restart)
        restart_server
        ;;
    status)
        status_server
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
