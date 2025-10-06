#!/bin/bash

# Demo script for Production Onboarding Graph
# This script demonstrates how to use the production onboarding system

echo "🚀 Production Onboarding Demo"
echo "=============================="
echo ""

# Set the API base URL
API_BASE="http://localhost:8080/api/v1"

echo "1. 📋 Listing available graphs..."
echo "--------------------------------"
curl -s "$API_BASE/graphs" | jq -r '.[] | "• \(.name) - \(.description)"'
echo ""

echo "2. 🎯 Getting Production Onboarding Graph details..."
echo "---------------------------------------------------"
PRODUCTION_GRAPH_ID=$(curl -s "$API_BASE/graphs" | jq -r '.[] | select(.name == "Production Onboarding") | .id')
echo "Production Graph ID: $PRODUCTION_GRAPH_ID"
echo ""

echo "3. 🚀 Starting a new onboarding session..."
echo "----------------------------------------"
SESSION_RESPONSE=$(curl -s -X POST "$API_BASE/sessions" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "demo-user", "graph_id": "'$PRODUCTION_GRAPH_ID'"}')
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r '.id')
echo "Session ID: $SESSION_ID"
echo ""

echo "4. 📍 Getting current node (Business Type Selection)..."
echo "------------------------------------------------------"
CURRENT_NODE=$(curl -s "$API_BASE/sessions/$SESSION_ID/current")
NODE_NAME=$(echo "$CURRENT_NODE" | jq -r '.name')
NODE_DESC=$(echo "$CURRENT_NODE" | jq -r '.description')
echo "Current Node: $NODE_NAME"
echo "Description: $NODE_DESC"
echo ""

echo "5. 📝 Available business types:"
echo "-------------------------------"
echo "$CURRENT_NODE" | jq -r '.fields[0].options[]' | nl -w2 -s'. '
echo ""

echo "6. 🎯 Submitting business type selection (Private Limited)..."
echo "-----------------------------------------------------------"
SUBMIT_RESPONSE=$(curl -s -X POST "$API_BASE/sessions/$SESSION_ID/submit" \
  -H "Content-Type: application/json" \
  -d '{"business_type": "private_limited"}')
echo "Submit Response:"
echo "$SUBMIT_RESPONSE" | jq '.'
echo ""

echo "7. 📍 Getting next node after business type selection..."
echo "------------------------------------------------------"
NEXT_NODE=$(curl -s "$API_BASE/sessions/$SESSION_ID/current")
NEXT_NODE_NAME=$(echo "$NEXT_NODE" | jq -r '.name')
NEXT_NODE_DESC=$(echo "$NEXT_NODE" | jq -r '.description')
echo "Next Node: $NEXT_NODE_NAME"
echo "Description: $NEXT_NODE_DESC"
echo ""

echo "8. 📋 Required fields for next node:"
echo "-----------------------------------"
echo "$NEXT_NODE" | jq -r '.fields[] | select(.required == true) | "• \(.name) (\(.type))"' 2>/dev/null || echo "No required fields found"
echo ""

echo "9. 🔍 Optional fields for next node:"
echo "-----------------------------------"
echo "$NEXT_NODE" | jq -r '.fields[] | select(.required == false) | "• \(.name) (\(.type))"' 2>/dev/null || echo "No optional fields found"
echo ""

echo "10. 📊 Session progress:"
echo "-----------------------"
curl -s "$API_BASE/sessions/$SESSION_ID" | jq -r '"Status: " + .status + " | Current Node: " + .current_node_id'
echo ""

echo "✅ Demo completed successfully!"
echo ""
echo "🌐 To test the full UI, open: http://localhost:8080/dynamic-onboarding-ui.html"
echo "   and select the 'Production Onboarding' graph to start the demo."
echo ""
echo "📚 Business Type Requirements:"
echo "• Individual: Personal PAN and basic documents"
echo "• Proprietorship: Personal PAN + MSME document"
echo "• Private Limited: Business PAN + CIN + Certificate of Incorporation"
echo "• Public Limited: Business PAN + CIN + Certificate of Incorporation"
echo "• Partnership: Business PAN + Partnership Deed"
echo "• LLP: Business PAN + Certificate of Incorporation"
echo "• Trust: Business PAN + Trust Deed"
echo "• Society: Business PAN + Society Registration Certificate"
echo "• HUF: Business PAN + HUF Deed"

