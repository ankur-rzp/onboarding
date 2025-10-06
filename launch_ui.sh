#!/bin/bash

# Simple UI Launcher for Production Onboarding System
# This script starts the backend and opens the UI

echo "🚀 Launching Production Onboarding System"
echo "========================================"

# Check if backend is already running
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Backend is already running on port 8080"
else
    echo "🔧 Starting backend server..."
    cd "$(dirname "$0")"
    go run main.go &
    BACKEND_PID=$!
    
    # Wait for backend to start
    echo "⏳ Waiting for backend to start..."
    for i in {1..10}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo "✅ Backend started successfully"
            break
        fi
        sleep 1
    done
    
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "❌ Failed to start backend"
        exit 1
    fi
fi

# Check available graphs
echo ""
echo "📋 Available onboarding flows:"
curl -s http://localhost:8080/api/v1/graphs | jq -r '.[] | "• \(.name) - \(.description)"'

echo ""
echo "🌐 Opening UI in browser..."
echo ""

# Open the UI in the default browser
if command -v open > /dev/null; then
    # macOS
    open http://localhost:8080/dynamic-onboarding-ui.html
elif command -v xdg-open > /dev/null; then
    # Linux
    xdg-open http://localhost:8080/dynamic-onboarding-ui.html
elif command -v start > /dev/null; then
    # Windows
    start http://localhost:8080/dynamic-onboarding-ui.html
else
    echo "Please open http://localhost:8080/dynamic-onboarding-ui.html in your browser"
fi

echo "🎉 UI launched successfully!"
echo ""
echo "📱 User Interface: http://localhost:8080/dynamic-onboarding-ui.html"
echo "👨‍💼 Admin Dashboard: http://localhost:8080/admin-dashboard.html"
echo "🔌 Backend API: http://localhost:8080/api/v1/"
echo ""
echo "💡 To test the Production Onboarding:"
echo "   1. Select 'Production Onboarding' from the available graphs"
echo "   2. Choose your business type (e.g., Private Limited)"
echo "   3. Follow the dynamic flow based on your selection"
echo ""
echo "Press Ctrl+C to stop the backend server (if started by this script)"

