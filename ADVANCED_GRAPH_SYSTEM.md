# Advanced Graph-Based Onboarding System

## 🎯 **System Overview**

Based on your diagram (`graph_nodes.drawio`), I've implemented an advanced graph-based onboarding system that supports:

1. **Multiple Entry Points** - Any node can be a starting point
2. **Dynamic Path Limitation** - Rules determine which paths are available based on user inputs
3. **Activation Rules** - Secondary rules evaluate if a user can be marked as "Activated"
4. **Conditional Navigation** - Different paths for different business types (individual vs company)

## 🏗️ **Architecture Components**

### 1. **Advanced Graph Engine** (`internal/onboarding/advanced_engine.go`)

**Key Features:**
- **Activation Rules**: Define when a user can be marked as "Activated"
- **Node Rules**: Apply rules when specific nodes are visited
- **Path Limitation**: Dynamically disable/enable edges based on user inputs
- **Multiple Entry Points**: Support for starting from any node

**Core Types:**
```go
type ActivationRule struct {
    ID            string
    Name          string
    Description   string
    RequiredNodes []string    // Nodes that must be visited
    ExcludedNodes []string    // Nodes that should not be visited
    Conditions    []ActivationCondition
}

type NodeRule struct {
    ID         string
    NodeID     string
    RuleType   string    // "validation", "path_limitation", "activation_check"
    Conditions []ActivationCondition
    Actions    []RuleAction
}
```

### 2. **Advanced Service** (`internal/onboarding/advanced_service.go`)

**Key Features:**
- **Session Management**: Track visited nodes, disabled edges, activation status
- **Rule Processing**: Apply rules when nodes are visited
- **Activation Checking**: Evaluate if user can be activated
- **Path Determination**: Calculate available paths based on current state

### 3. **Example Implementation** (`examples/advanced_onboarding.go`)

**Graph Structure from Your Diagram:**
```
Start → [Email, Phone, User PAN, Business Type, GST, Company Cert]
  ↓
Phone → Business Type
  ↓
Business Type → GST (if not individual) → Company Cert → PAN → User PAN → Bank Account → Activated
  ↓
Business Type → Activated (if individual)
```

## 🔄 **Dynamic Flow Examples**

### **Individual User Flow:**
1. **Start** → **Phone** → **Business Type** (select "individual")
2. **Node Rule Applied**: Exclude GST and Company Cert nodes
3. **Available Paths**: Direct to Activated (if all required data collected)
4. **Activation Rule**: Individual users need Phone + User PAN + Bank Account

### **Company User Flow:**
1. **Start** → **Phone** → **Business Type** (select "company")
2. **Available Paths**: GST → Company Cert → PAN → User PAN → Bank Account → Activated
3. **Activation Rule**: Company users need all business-related nodes

## 🎯 **Key Features Implemented**

### 1. **Multiple Entry Points**
```go
// Any node can be a starting point
entryNodes := []string{
    "start", "email", "phone", "user_pan", 
    "business_type", "gst", "company_cert"
}
```

### 2. **Dynamic Path Limitation**
```go
// Business type rule excludes certain nodes for individuals
NodeRule{
    NodeID: "business_type",
    Conditions: []ActivationCondition{
        {
            Field:    "business_type",
            Operator: "eq",
            Value:    "individual",
        },
    },
    Actions: []RuleAction{
        {Type: "exclude_node", Target: "gst"},
        {Type: "exclude_node", Target: "company_cert"},
    },
}
```

### 3. **Activation Rules**
```go
// Individual user activation rule
ActivationRule{
    ID: "individual_activation",
    RequiredNodes: []string{"phone", "user_pan", "bank_account"},
    ExcludedNodes: []string{"gst", "company_cert"},
    Conditions: []ActivationCondition{
        {
            Field:    "business_type",
            Operator: "eq",
            Value:    "individual",
            Required: true,
        },
    },
}
```

### 4. **Conditional Edge Traversal**
```go
// GST edge only available for non-individual users
EdgeCondition{
    Type:     "field_value",
    Field:    "business_type",
    Operator: "ne",
    Value:    "individual",
}
```

## 🧪 **Testing**

The system includes comprehensive tests that verify:

1. **Activation Rules**: Different rules for individual vs company users
2. **Node Rules**: Path limitation based on business type
3. **Path Availability**: Dynamic path calculation based on current state
4. **Edge Conditions**: Conditional traversal based on user inputs

**Test Results:**
```
=== RUN   TestAdvancedEngine_ActivationRules
--- PASS: TestAdvancedEngine_ActivationRules (0.00s)
=== RUN   TestAdvancedEngine_NodeRules  
--- PASS: TestAdvancedEngine_NodeRules (0.00s)
=== RUN   TestAdvancedEngine_GetAvailablePaths
--- PASS: TestAdvancedEngine_GetAvailablePaths (0.00s)
```

## 🚀 **Usage Examples**

### **Starting a Session with Entry Point:**
```go
session, err := service.StartAdvancedSession(ctx, "user123", "graph456", "phone")
```

### **Submitting Data and Processing Rules:**
```go
result, err := service.SubmitAdvancedNodeData(ctx, sessionID, map[string]interface{}{
    "business_type": "individual",
    "phone": "9876543210",
})
```

### **Checking Activation Status:**
```go
activationResult, err := service.CheckActivationStatus(ctx, sessionID)
if activationResult.CanActivate {
    // User can be activated
}
```

## 🎉 **Benefits Achieved**

1. **Flexibility**: Any node can be an entry point
2. **Dynamic Paths**: Paths change based on user inputs
3. **Smart Activation**: Users can be activated through different paths
4. **Rule-Based**: Easy to add new rules and conditions
5. **Government Compliance**: Built-in support for different business types
6. **Fault Tolerance**: Comprehensive error handling and validation

## 🔮 **Future Enhancements**

1. **Rule Engine Integration**: Support for external rule engines (OPA, Drools)
2. **Visual Graph Editor**: UI for creating and editing graphs
3. **Analytics Dashboard**: Track completion rates and bottlenecks
4. **A/B Testing**: Test different graph configurations
5. **Machine Learning**: AI-powered path optimization

This advanced system successfully implements the sophisticated graph structure from your diagram, providing the flexibility and dynamic behavior you requested while maintaining fault tolerance and consistency.
