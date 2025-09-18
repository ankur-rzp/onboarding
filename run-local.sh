#!/bin/bash

echo "🚀 Starting Onboarding System in Local Mode"
echo "==========================================="

# Set environment variables for local development
export SERVER_ADDRESS=":8080"
export ONBOARDING_MAX_RETRIES=3
export ONBOARDING_RETRY_DELAY="5s"
export ONBOARDING_SESSION_TIMEOUT="24h"

# Don't set database variables - this will trigger in-memory storage
unset DB_HOST
unset DB_USER
unset DB_PASSWORD
unset DB_NAME
unset REDIS_HOST

echo "📝 Configuration:"
echo "   - Server: $SERVER_ADDRESS"
echo "   - Storage: In-Memory (no database required)"
echo "   - Max Retries: $ONBOARDING_MAX_RETRIES"
echo "   - Session Timeout: $ONBOARDING_SESSION_TIMEOUT"
echo ""

echo "🔧 Building application..."
go build -o onboarding-system-local .

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"
echo ""

echo "🎯 Starting server..."
echo "   - Health check: http://localhost:8080/health"
echo "   - API docs: See README.md"
echo "   - Press Ctrl+C to stop"
echo ""

# Start the application
./onboarding-system-local
