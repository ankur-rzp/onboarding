#!/bin/bash

# Quick Launch Script for Production Onboarding System
# This script opens the UI directly and shows the current status

echo "🚀 Quick Launch - Production Onboarding System"
echo "=============================================="

# Check if backend is running
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Backend is running on port 8080"
    
    # Show available graphs
    echo ""
    echo "📋 Available onboarding flows:"
    curl -s http://localhost:8080/api/v1/graphs | jq -r '.[] | "• \(.name) - \(.description)"'
    
    echo ""
    echo "🌐 Opening UI in browser..."
    
    # Open the HTML file directly
    if command -v open > /dev/null; then
        # macOS
        open /Users/ankur.gogate/projects/poc/onboarding_grabh_based/dynamic-onboarding-ui.html
    elif command -v xdg-open > /dev/null; then
        # Linux
        xdg-open /Users/ankur.gogate/projects/poc/onboarding_grabh_based/dynamic-onboarding-ui.html
    elif command -v start > /dev/null; then
        # Windows
        start /Users/ankur.gogate/projects/poc/onboarding_grabh_based/dynamic-onboarding-ui.html
    else
        echo "Please open the file: /Users/ankur.gogate/projects/poc/onboarding_grabh_based/dynamic-onboarding-ui.html"
    fi
    
    echo ""
    echo "🎉 UI launched successfully!"
    echo ""
    echo "💡 To test the Production Onboarding:"
    echo "   1. Select 'Production Onboarding' from the available graphs"
    echo "   2. Choose your business type (e.g., Private Limited)"
    echo "   3. Follow the dynamic flow based on your selection"
    echo ""
    echo "🔌 Backend API: http://localhost:8080/api/v1/"
    
else
    echo "❌ Backend is not running on port 8080"
    echo ""
    echo "Please start the backend first:"
    echo "   cd /Users/ankur.gogate/projects/poc/onboarding_grabh_based"
    echo "   go run main.go"
    echo ""
    echo "Or use the full startup script:"
    echo "   ./start_services.sh"
fi
