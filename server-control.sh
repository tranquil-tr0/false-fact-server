#!/bin/bash

# script for false fact backend/api/server control
# Usage: ./server-control.sh [start|stop|restart|status] [port]

SERVER_NAME="server"
DEFAULT_PORT="3088"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PID_FILE="$SCRIPT_DIR/.${SERVER_NAME}.pid"
LOG_FILE="$SCRIPT_DIR/${SERVER_NAME}.log"

# Allow port to be specified as second argument
PORT=${2:-$DEFAULT_PORT}

start_server() {
    if [ -f "$PID_FILE" ]; then
        echo "Server is already running (PID: $(cat $PID_FILE))"
        return 1
    fi
    
    # Check if port is available
    if netstat -tuln | grep -q ":$PORT "; then
        echo "Error: Port $PORT is already in use!"
        echo "Try a different port: ./deploy.sh start [port_number]"
        return 1
    fi
    
    echo "Starting $SERVER_NAME on port $PORT..."
    export PORT=$PORT
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

# Function to find an available port
find_available_port() {
    local start_port=${1:-3088}
    local port=$start_port
    
    while netstat -tuln | grep -q ":$port "; do
        port=$((port + 1))
        if [ $port -gt 65535 ]; then
            echo "Error: No available ports found"
            return 1
        fi
    done
    
    echo $port
}

case "$1" in
    start)
        # If no port specified and default is taken, find available port
        if [ -z "$2" ] && netstat -tuln | grep -q ":$DEFAULT_PORT "; then
            AVAILABLE_PORT=$(find_available_port $DEFAULT_PORT)
            echo "Port $DEFAULT_PORT is in use, trying port $AVAILABLE_PORT"
            PORT=$AVAILABLE_PORT
        fi
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
        echo "Usage: $0 {start|stop|restart|status} [port]"
        echo "Examples:"
        echo "  $0 start          # Start on default port (3000) or find available"
        echo "  $0 start 3001     # Start on specific port"
        echo "  $0 restart 3002   # Restart on specific port"
        exit 1
        ;;
esac
