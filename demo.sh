#!/bin/bash

echo "🚀 Dynamic Onboarding System Demo"
echo "=================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

echo "📦 Starting the system with Docker Compose..."
docker-compose up -d

echo "⏳ Waiting for services to be ready..."
sleep 10

echo "🔍 Checking system health..."
curl -s http://localhost:8080/health | jq .

echo ""
echo "📊 Creating example onboarding flows..."
# Note: In a real scenario, you would use the seed command here
# ./seed-examples

echo ""
echo "🎯 Example API Usage:"
echo ""
echo "1. List available graphs:"
echo "   curl http://localhost:8080/api/v1/graphs"
echo ""
echo "2. Start a new session:"
echo "   curl -X POST http://localhost:8080/api/v1/sessions \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"user_id\": \"demo-user\", \"graph_id\": \"your-graph-id\"}'"
echo ""
echo "3. Get current node:"
echo "   curl http://localhost:8080/api/v1/sessions/{session_id}/current"
echo ""
echo "4. Submit data:"
echo "   curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/submit \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"field_name\": \"value\"}'"
echo ""
echo "5. Go back:"
echo "   curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/back"
echo ""

echo "🏗️  System Architecture:"
echo "   - Graph-based onboarding flows"
echo "   - Dynamic validation rules"
echo "   - Fault tolerance with retry mechanisms"
echo "   - PostgreSQL for persistence"
echo "   - Redis for caching"
echo "   - RESTful API"
echo ""

echo "✅ System is ready! Visit http://localhost:8080/health to check status"
echo "📚 See README.md for detailed usage instructions"
echo ""
echo "🛑 To stop the system: docker-compose down"
