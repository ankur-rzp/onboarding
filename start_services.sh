#!/bin/bash

# Onboarding System Startup Script
# This script starts both the backend and frontend services

echo "🚀 Starting Onboarding System Services..."
echo "=================================="

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH. Exiting."
    exit 1
fi

echo "✅ Go is available"

# Function to cleanup background processes
cleanup() {
    echo ""
    echo "🛑 Shutting down services..."
    pkill -f "go run main.go" 2>/dev/null
    pkill -f "python3 -m http.server 3000" 2>/dev/null
    echo "✅ Services stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Start backend in background
echo "🔧 Starting backend server on port 8080..."
go run main.go &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 3

# Check if backend started successfully
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo "❌ Backend failed to start. Exiting."
    exit 1
fi

echo "✅ Backend started successfully (PID: $BACKEND_PID)"

# Start frontend
echo "🌐 Starting frontend server on port 3000..."
echo ""
echo "🎉 Services are running!"
echo "=================================="
echo "📱 User UI: http://localhost:3000/dynamic-onboarding-ui.html"
echo "👨‍💼 Admin Dashboard: http://localhost:3000/admin-dashboard.html"
echo "🔌 Backend API: http://localhost:8080/api/v1/"
echo "=================================="
echo ""
echo "Press Ctrl+C to stop all services"

# Start frontend (this will block)
python3 -m http.server 3000

