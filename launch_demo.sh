#!/bin/bash

echo "🚀 Launching Dynamic Onboarding Demo"
echo "===================================="

# Check if backend is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "⚠️  Backend not running. Starting backend..."
    
    # Kill any existing backend processes
    pkill -f "go run main.go" 2>/dev/null || true
    sleep 2
    
    # Start backend in background
    echo "🔧 Starting backend server on port 8080..."
    go run main.go &
    BACKEND_PID=$!
    
    # Wait for backend to start
    echo "⏳ Waiting for backend to start..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo "✅ Backend started successfully"
            break
        fi
        sleep 1
        echo -n "."
    done
    
    if [ $i -eq 30 ]; then
        echo "❌ Backend failed to start"
        kill $BACKEND_PID 2>/dev/null || true
        exit 1
    fi
else
    echo "✅ Backend is already running"
fi

# Open the demo UI
echo "🌐 Opening Dynamic Onboarding Demo UI..."
echo "📱 Demo URL: file://$(pwd)/dynamic-onboarding-demo.html"

# Try to open in browser
if command -v open > /dev/null; then
    open dynamic-onboarding-demo.html
elif command -v xdg-open > /dev/null; then
    xdg-open dynamic-onboarding-demo.html
elif command -v start > /dev/null; then
    start dynamic-onboarding-demo.html
else
    echo "⚠️  Could not auto-open browser. Please open:"
    echo "   file://$(pwd)/dynamic-onboarding-demo.html"
fi

echo ""
echo "🎉 Demo launched successfully!"
echo ""
echo "📋 Demo Features:"
echo "  ✅ Business type selection with dynamic requirements"
echo "  ✅ Real-time node status updates"
echo "  ✅ Session persistence and reload simulation"
echo "  ✅ State summary and validation"
echo "  ✅ Activity logging"
echo ""
echo "🔧 Backend API: http://localhost:8080"
echo "📱 Demo UI: file://$(pwd)/dynamic-onboarding-demo.html"
echo ""
echo "Press Ctrl+C to stop the backend when done"
