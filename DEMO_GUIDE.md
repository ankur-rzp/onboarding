# 🚀 Dynamic Onboarding System - Demo Guide

## 📋 Overview

This demo showcases the **Dynamic Onboarding System** with observer pattern and persistence, demonstrating how CSV-based business type requirements are handled with real-time node status updates.

## 🎯 Demo Features

### ✅ **Core Functionality**
- **Business Type Selection**: Choose from 9 different business types (Individual, Proprietorship, Private Limited, etc.)
- **Dynamic Node Status**: Real-time updates showing mandatory, optional, dependent, and completed nodes
- **Session Persistence**: State survives server restarts and session reloads
- **Observer Pattern**: Automatic status changes based on user actions
- **State Validation**: Comprehensive validation and error reporting

### ✅ **UI Components**
- **Session Information**: Shows current session ID, business type, and completion status
- **Business Type Cards**: Interactive selection with visual feedback
- **Node Status List**: Color-coded status indicators with dependency counts
- **Activity Log**: Real-time logging of all actions and state changes
- **Session Data**: Live view of all collected user data

## 🚀 Getting Started

### 1. **Launch the Demo**
```bash
# Make sure you're in the project directory
cd /path/to/onboarding_grabh_based

# Launch the demo (starts backend + opens UI)
./launch_demo.sh
```

### 2. **Alternative Manual Launch**
```bash
# Start backend
go run main.go

# Open the demo UI in your browser
open dynamic-onboarding-demo.html
# OR
file:///path/to/onboarding_grabh_based/dynamic-onboarding-demo.html
```

## 📱 Demo Walkthrough

### **Step 1: Start a Session**
1. Click **"🆕 Start New Session"** button
2. Watch the activity log for session creation
3. Note the session ID in the header

### **Step 2: Select Business Type**
1. Click on any business type card (e.g., "Individual", "Proprietorship")
2. Observe the real-time updates:
   - Business type updates in header
   - Node statuses change dynamically
   - Activity log shows the selection

### **Step 3: Observe Dynamic Behavior**
1. **Node Status Colors**:
   - 🔴 **Red (Mandatory)**: Required for completion
   - 🔵 **Blue (Optional)**: Not required but available
   - 🟡 **Yellow (Dependent)**: Becomes mandatory based on conditions
   - 🟢 **Green (Completed)**: User has filled this node
   - ⚪ **Gray (Disabled)**: Not available for current flow

2. **Dependency Counts**: Each node shows how many dependencies it has

### **Step 4: Test Persistence**
1. Click **"🔄 Simulate Reload"** button
2. Observe that:
   - New session is created
   - Previous data is preserved
   - Node statuses are restored
   - Business type selection is maintained

### **Step 5: View State Summary**
1. Click **"📊 Show State Summary"** button
2. Review the detailed state information:
   - Total nodes and their status counts
   - Validation issues (if any)
   - Business type consistency

## 🏢 Business Type Requirements (From CSV)

### **Individual**
- **Mandatory**: Business Type Selection, PAN Number, Payment Channel, Business Information
- **Optional**: MCC & Policy Verification, BMC Document, Bank Account Details
- **Dependent**: Authorised Signatory Details, Business Document

### **Proprietorship**
- **Mandatory**: All Individual nodes + MCC & Policy Verification, Bank Account Details
- **Dependent**: Business Document (MSME document required)

### **Private Limited**
- **Mandatory**: All Proprietorship nodes + Authorised Signatory Details, Business Document
- **Dependent**: None (all requirements are mandatory)

### **Partnership**
- **Mandatory**: All Private Limited nodes
- **Dependent**: Business Document (Partnership deed required)

### **LLP, Trust, Society, HUF**
- Similar patterns with business-type specific requirements

## 🔍 Key Observations

### **1. Dynamic Status Changes**
- Nodes change from "dependent" to "mandatory" when business type is selected
- Status updates happen in real-time without page refresh
- Observer pattern ensures all dependent nodes are notified

### **2. Persistence Behavior**
- Session data is automatically saved after each action
- State survives server restarts
- Business type changes are persisted and recalculated

### **3. Validation & Error Handling**
- Real-time validation feedback
- Clear error messages in activity log
- State consistency checks

## 🛠️ Technical Features Demonstrated

### **Observer Pattern**
- Real-time notifications when user data changes
- Automatic re-evaluation of dependent nodes
- Clean separation of concerns

### **Persistence Layer**
- Dynamic state serialization/deserialization
- Session state restoration
- Business type consistency validation

### **API Integration**
- RESTful API calls to dynamic endpoints
- Real-time data synchronization
- Error handling and user feedback

## 📊 API Endpoints Used

```bash
# Start dynamic session
POST /api/v1/dynamic/sessions

# Update business type
PUT /api/v1/dynamic/sessions/{id}/business-type

# Submit node data
POST /api/v1/dynamic/sessions/{id}/submit

# Get node status
GET /api/v1/dynamic/sessions/{id}/status

# Get state summary
GET /api/v1/dynamic/sessions/{id}/summary
```

## 🎯 Demo Scenarios

### **Scenario 1: Basic Flow**
1. Start session → Select "Individual" → Observe node statuses
2. Change to "Proprietorship" → See additional mandatory nodes
3. Change to "Private Limited" → See all nodes become mandatory

### **Scenario 2: Persistence Test**
1. Start session → Select business type → Submit some data
2. Simulate reload → Verify data is preserved
3. Check state summary → Verify consistency

### **Scenario 3: Business Type Comparison**
1. Start multiple sessions with different business types
2. Compare node statuses side by side
3. Observe how requirements change dynamically

## 🔧 Troubleshooting

### **Backend Not Starting**
```bash
# Check if port 8080 is in use
lsof -i :8080

# Kill existing processes
pkill -f "go run main.go"

# Start fresh
go run main.go
```

### **UI Not Loading**
- Ensure backend is running on port 8080
- Check browser console for errors
- Verify file path is correct

### **API Errors**
- Check activity log for detailed error messages
- Verify session is active before making requests
- Ensure all required fields are provided

## 📈 Performance Notes

- **Real-time Updates**: Node statuses update instantly
- **Persistence**: State is saved after each action
- **Scalability**: System handles multiple concurrent sessions
- **Memory**: In-memory storage for demo (production would use database)

## 🎉 Success Criteria

The demo is working correctly when you can:

1. ✅ Start a new session successfully
2. ✅ Select different business types and see node status changes
3. ✅ Observe real-time updates in the activity log
4. ✅ Simulate session reload and see data persistence
5. ✅ View comprehensive state summaries
6. ✅ See proper error handling and validation

## 🚀 Next Steps

After the demo, you can:

1. **Integrate with Production**: Replace in-memory storage with database
2. **Add More Business Types**: Extend the CSV requirements
3. **Customize UI**: Modify the demo UI for your specific needs
4. **Add Authentication**: Implement user authentication and authorization
5. **Deploy**: Deploy to production environment

---

## 📞 Support

If you encounter any issues:

1. Check the activity log for error messages
2. Verify backend is running on port 8080
3. Check browser console for JavaScript errors
4. Review the demo guide for troubleshooting steps

**Happy Demo-ing! 🎉**
