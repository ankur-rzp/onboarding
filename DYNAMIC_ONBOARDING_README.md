# Dynamic Onboarding System with Observer Pattern

## 🎯 Overview

The Dynamic Onboarding System is a revolutionary approach to managing onboarding flows that eliminates the need for complex static rule groups. Instead, it uses an **observer pattern** where nodes dynamically change their status (mandatory/optional/dependent) based on user actions and business type requirements.

## 🚀 Key Features

### ✅ **Dynamic Node Status Management**
- Nodes automatically determine their status based on business type and user actions
- Real-time status updates as users fill forms and make selections
- No need for complex static rule definitions

### ✅ **Observer Pattern Implementation**
- Real-time notifications when user data changes
- Automatic re-evaluation of dependent nodes
- Clean separation of concerns

### ✅ **Business Type Specific Logic**
- Each business type has its own set of mandatory/optional nodes
- Dynamic requirements based on user selections
- Flexible and maintainable architecture

### ✅ **Smart Dependency Resolution**
- Nodes become mandatory when their dependencies are satisfied
- Conditional logic based on user input
- Automatic status transitions

### ✅ **Persistent State Management**
- Dynamic node statuses are saved to the database
- Session state survives server restarts and reloads
- Automatic state restoration on session reload
- Business type changes are persisted and recalculated

## 🏗️ Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Dynamic Onboarding System                │
├─────────────────────────────────────────────────────────────┤
│  DynamicEngine  │  DynamicService  │  DynamicHandlers       │
│                 │                  │                        │
│  • Status Logic │  • Session Mgmt  │  • API Endpoints       │
│  • Dependencies │  • Node Updates  │  • Request Handling    │
│  • Observer     │  • Validation    │  • Response Formatting │
│  • Persistence  │  • State Mgmt    │  • CORS Support        │
└─────────────────────────────────────────────────────────────┘
│                                                             │
│  DynamicGraph  │  DynamicNode  │  PersistenceManager       │
│                │               │                           │
│  • Graph Mgmt  │  • Status     │  • State Serialization    │
│  • Observers   │  • Dependencies│  • Session Restoration   │
│  • Updates     │  • Validation │  • State Validation       │
└─────────────────────────────────────────────────────────────┘
```

### Node Status Types

| Status | Description | Example |
|--------|-------------|---------|
| `mandatory` | Required for completion | Business Type Selection, PAN Number |
| `optional` | Not required but available | BMC Document, Additional Info |
| `dependent` | Becomes mandatory based on conditions | Business Document (depends on business type) |
| `completed` | User has filled this node | Any node after user submission |
| `disabled` | Not available for current flow | Irrelevant nodes for business type |

## 🔄 How It Works

### 1. **Initialization Phase**
```go
// Create dynamic graph for business type
dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

// Nodes get initial status based on business type
// - Business Type Selection: mandatory
// - PAN Number: mandatory  
// - Payment Channel: mandatory
// - Business Document: dependent (7 dependencies)
```

### 2. **User Action Phase**
```go
// User selects business type
sessionData["business_type"] = "individual"
dynamicGraph.OnNodeCompleted("business_type_selection", sessionData)

// Observer pattern triggers re-evaluation
// Dependent nodes are automatically updated
```

### 3. **Dynamic Status Updates**
```go
// When user changes payment channel to "app"
sessionData["payment_channel"] = "app"
dynamicGraph.OnNodeDataChanged("payment_channel", "payment_channel", "app", sessionData)

// System automatically evaluates dependencies
// Updates node statuses in real-time
```

### 4. **Completion Check**
```go
completionStatus := dynamicGraph.GetCompletionStatus()
// Returns: {
//   "total_nodes": 10,
//   "mandatory_nodes": 1,
//   "completed_nodes": 3,
//   "can_complete": false
// }
```

### 5. **Persistence & Reload**
```go
// State is automatically saved after each node submission
ds.persistenceManager.SaveDynamicState(session, dynamicGraph, businessType)

// On session reload, state is automatically restored
ds.persistenceManager.RestoreDynamicState(session, dynamicGraph)

// Session survives server restarts and maintains all dynamic states
```

## 📊 Business Type Requirements

### Individual
- **Mandatory**: Business Type Selection, PAN Number, Payment Channel, Business Information
- **Optional**: MCC & Policy Verification, BMC Document, Bank Account Details
- **Dependent**: Authorised Signatory Details, Business Document

### Proprietorship
- **Mandatory**: All Individual nodes + MCC & Policy Verification, Bank Account Details
- **Dependent**: Business Document (MSME document required)

### Private Limited
- **Mandatory**: All Proprietorship nodes + Authorised Signatory Details, Business Document
- **Dependent**: None (all requirements are mandatory)

### Partnership
- **Mandatory**: All Private Limited nodes
- **Dependent**: Business Document (Partnership deed required)

## 🛠️ Implementation

### Creating a Dynamic Graph

```go
// Initialize dynamic engine
dynamicEngine := onboarding.NewDynamicEngine(logger)

// Convert regular graph to dynamic graph
dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

// Graph now has dynamic node management
```

### Handling User Actions

```go
// User submits node data
result, err := dynamicService.SubmitNodeDataDynamic(ctx, sessionID, data)

// System automatically:
// 1. Validates the data
// 2. Updates session data
// 3. Marks node as completed
// 4. Notifies observers
// 5. Re-evaluates dependent nodes
// 6. Determines next node
```

### Getting Node Status

```go
// Get current status of all nodes
status, err := dynamicService.GetDynamicNodeStatus(ctx, sessionID)

// Returns detailed status for each node:
// {
//   "nodes": {
//     "node_id": {
//       "name": "PAN Number",
//       "status": "completed",
//       "initial_status": "mandatory",
//       "dependencies": [...]
//     }
//   }
// }
```

## 🔌 API Endpoints

### Dynamic Session Management

```bash
# Start dynamic session
POST /api/v1/dynamic/sessions
{
  "graph_id": "graph-uuid",
  "user_id": "user-123"
}

# Submit node data with dynamic updates
POST /api/v1/dynamic/sessions/{id}/submit
{
  "business_type": "individual",
  "pan_number": "ABCDE1234F"
}

# Get dynamic node status
GET /api/v1/dynamic/sessions/{id}/status

# Update business type
PUT /api/v1/dynamic/sessions/{id}/business-type
{
  "business_type": "proprietorship"
}

# Get eligible nodes
GET /api/v1/dynamic/sessions/{id}/eligible-nodes

# Get dynamic node status (with persistence)
GET /api/v1/dynamic/sessions/{id}/status

# Update business type and recalculate
PUT /api/v1/dynamic/sessions/{id}/business-type
{
  "business_type": "proprietorship"
}

# Get dynamic state summary
GET /api/v1/dynamic/sessions/{id}/summary
```

## 🧪 Testing

### Running Tests

```bash
# Run all dynamic system tests
go test ./examples -v -run TestDynamicOnboardingSystem

# Run specific test
go test ./examples -v -run TestDynamicOnboardingSystem/ConvertToDynamicGraph

# Run benchmarks
go test ./examples -bench=BenchmarkDynamicGraphConversion

# Run persistence tests
go test ./examples -v -run TestDynamicPersistence

# Run persistence performance tests
go test ./examples -v -run TestDynamicPersistencePerformance
```

### Demo

```bash
# Run the comprehensive demo
go run examples/dynamic_onboarding_demo.go

# Run the persistence demo
go run demo_persistence.go

# This will show:
# - Initial node statuses
# - User action simulation
# - Dynamic status changes
# - Business type comparisons
# - State persistence and restoration
# - Session reload simulation
```

## 📈 Benefits Over Static Rule Groups

| Aspect | Static Rule Groups | Dynamic System |
|--------|-------------------|----------------|
| **Complexity** | High - complex rule definitions | Low - simple status tracking |
| **Maintainability** | Difficult - rules scattered | Easy - centralized logic |
| **Flexibility** | Limited - predefined paths | High - dynamic adaptation |
| **Performance** | Slow - complex evaluations | Fast - simple status checks |
| **Debugging** | Hard - rule interactions | Easy - clear status flow |
| **Extensibility** | Difficult - new rules needed | Easy - new business types |

## 🔧 Configuration

### Adding New Business Types

```go
// In dynamic_engine.go, update businessTypeRequirements map
businessTypeRequirements := map[string][]string{
    "new_business_type": {
        "Required Node 1",
        "Required Node 2",
        // ... more required nodes
    },
}
```

### Adding New Dependencies

```go
// Dependencies are automatically extracted from node validation conditions
// Just add conditions to your node definition:

node.Validation.Conditions = []types.ValidationCondition{
    {
        Field:    "business_type",
        Operator: "eq",
        Value:    "new_type",
        Rule:     "Special document required for new business type",
    },
}
```

## 🚀 Getting Started

### 1. **Basic Usage**

```go
package main

import (
    "onboarding-system/examples"
)

func main() {
    // Run the demo to see the system in action
    examples.RunDynamicDemo()
}
```

### 2. **Integration with Existing System**

```go
// Replace regular service with dynamic service
dynamicService := onboarding.NewDynamicService(storage, config, logger)

// Use dynamic endpoints
dynamicHandlers := api.NewDynamicHandlers(dynamicService, logger)
dynamicHandlers.RegisterDynamicRoutes(router)
```

### 3. **UI Integration**

```javascript
// Frontend can now get real-time node status
const response = await fetch('/api/v1/dynamic/sessions/' + sessionId + '/status');
const status = await response.json();

// Update UI based on node status
status.nodes.forEach(node => {
    updateNodeStatus(node.id, node.status);
});
```

## 📝 Example Flow

### User Journey: Individual Business

1. **Start**: User selects "Individual" business type
   - Business Type Selection: `mandatory` → `completed`
   - PAN Number: `mandatory`
   - Payment Channel: `mandatory`
   - Business Information: `mandatory`

2. **PAN Entry**: User fills PAN number
   - PAN Number: `mandatory` → `completed`
   - System checks dependencies

3. **Payment Channel**: User selects "Website"
   - Payment Channel: `mandatory` → `completed`
   - Website URL field becomes required

4. **Business Info**: User fills business details
   - Business Information: `mandatory` → `completed`
   - All mandatory nodes completed!

5. **Completion**: System allows user to complete onboarding
   - Status: `can_complete: true`

## 🔍 Debugging

### Enable Debug Logging

```go
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)

// This will show detailed logs of:
// - Node status changes
// - Dependency evaluations
// - Observer notifications
// - Completion status updates
```

### Common Issues

1. **Node not becoming mandatory**: Check dependencies and business type requirements
2. **Status not updating**: Verify observer pattern is working correctly
3. **Completion not detected**: Ensure all mandatory nodes are completed

## 🤝 Contributing

### Adding New Features

1. **New Node Status**: Add to `NodeStatus` enum
2. **New Dependencies**: Extend `NodeDependency` struct
3. **New Business Types**: Update `businessTypeRequirements` map
4. **New Observers**: Implement `NodeObserver` interface

### Code Style

- Follow Go conventions
- Add comprehensive tests
- Update documentation
- Include examples

## 📚 References

- [Observer Pattern](https://en.wikipedia.org/wiki/Observer_pattern)
- [Go Concurrency Patterns](https://golang.org/doc/effective_go.html#concurrency)
- [Dynamic Programming](https://en.wikipedia.org/wiki/Dynamic_programming)

---

## 🎉 Conclusion

The Dynamic Onboarding System represents a significant improvement over traditional static rule-based approaches. By leveraging the observer pattern and dynamic status management, it provides:

- **Simplified Architecture**: No complex rule definitions
- **Real-time Updates**: Immediate response to user actions  
- **Flexible Requirements**: Easy to add new business types
- **Better Performance**: Simple status checks vs complex rule evaluation
- **Maintainable Code**: Clear separation of concerns

This system is production-ready and can handle complex onboarding flows with ease! 🚀
