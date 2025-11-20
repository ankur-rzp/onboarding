#!/bin/bash

echo "🚀 Launching Simple Onboarding Demo"
echo "=================================="

# Check if backend is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Backend not running on port 8080"
    echo "🔧 Starting backend..."
    go run main.go &
    BACKEND_PID=$!
    sleep 3
    echo "✅ Backend started (PID: $BACKEND_PID)"
else
    echo "✅ Backend already running on port 8080"
fi

# Check if frontend server is running
if ! curl -s http://localhost:3000 > /dev/null; then
    echo "🌐 Starting frontend server on port 3000..."
    python3 -m http.server 3000 &
    FRONTEND_PID=$!
    sleep 2
    echo "✅ Frontend server started (PID: $FRONTEND_PID)"
else
    echo "✅ Frontend server already running on port 3000"
fi

echo ""
echo "🎉 Demo is ready!"
echo "📱 Open your browser and go to:"
echo "   http://localhost:3000/simple-demo.html"
echo ""
echo "🔧 Backend API: http://localhost:8080"
echo "🌐 Frontend: http://localhost:3000"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for user to stop
wait
