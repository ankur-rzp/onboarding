# 🚀 Dynamic Onboarding System - Complete Implementation Guide

## Table of Contents
1. [Overview](#overview)
2. [Node Configuration & Graph Creation](#1-node-configuration--graph-creation)
3. [Edge Configuration & Dynamic Navigation](#2-edge-configuration--dynamic-navigation)
4. [Node Validation Logic on Save](#3-node-validation-logic-on-save)
5. [Observer Pattern for Dynamic Node Status](#4-observer-pattern-for-dynamic-node-status)
6. [Rule Group Validation for Business Type Completion](#5-rule-group-validation-for-business-type-completion)
7. [Cross-Node Validation Integration](#6-cross-node-validation-integration)
8. [Complete Validation Flow](#complete-validation-flow)
9. [Key Benefits](#key-benefits)
10. [API Endpoints](#api-endpoints)
11. [Testing](#testing)
12. [Architecture Diagrams](#architecture-diagrams)

## Overview

The Dynamic Onboarding System is a flexible, rule-based onboarding flow that adapts based on user input and business requirements. It supports:

- **Dynamic Navigation**: Users can fill forms in any order
- **Business Type Specific Rules**: Different validation rules for different business types
- **Cross-Node Validation**: Data consistency across multiple nodes
- **Observer Pattern**: Real-time status updates for dependent nodes
- **Rule Group System**: Multiple paths to completion
- **Persistence**: State maintained across sessions

## 1. Node Configuration & Graph Creation

### Node Structure with Dynamic Metadata

Each node is configured with dynamic properties that determine its behavior in the onboarding flow:

```go
// Each node is configured with dynamic properties
startNode := &types.Node{
    ID:          "business_type_selection",
    Type:        "start",
    Name:        "Business Type Selection",
    Description: "Select your business type",
    Fields: []types.Field{
        {
            ID:       "business_type",
            Name:     "business_type", 
            Type:     types.FieldTypeSelect,
            Required: true,
            Options:  []string{"individual", "proprietorship", "private_limited", ...},
        },
    },
    // Dynamic properties
    IsIndependent: true,    // Can be accessed from start node
    IsDependent:   false,   // Doesn't require dependencies
    Dependencies:  []types.NodeDependency{}, // No dependencies
    IncomingEdges: []string{}, // Will be populated when edges are built
    OutgoingEdges: []string{}, // Will be populated when edges are built
}
```

### Node Categories

The system categorizes nodes into different types:

- **🟢 Start Node**: Entry point (Business Type Selection)
- **🔵 Independent Nodes**: Always accessible (PAN, Payment Channel, MCC & Policy, Business Info)
- **🟡 Dependent Nodes**: Require conditions (Business Document, BMC Document, Authorised Signatory, Bank Account)
- **🟣 Completion Node**: End point (Onboarding Complete)

### Field Types

```go
const (
    FieldTypeText     FieldType = "text"
    FieldTypeEmail    FieldType = "email"
    FieldTypeNumber   FieldType = "number"
    FieldTypeSelect   FieldType = "select"
    FieldTypeFile     FieldType = "file"
    FieldTypeCheckbox FieldType = "checkbox"
    FieldTypeDate     FieldType = "date"
)
```

### Field Validation

```go
type FieldValidation struct {
    Pattern     string   `json:"pattern,omitempty"`
    MinLength   int      `json:"min_length,omitempty"`
    MaxLength   int      `json:"max_length,omitempty"`
    MinValue    float64  `json:"min_value,omitempty"`
    MaxValue    float64  `json:"max_value,omitempty"`
    CustomRules []string `json:"custom_rules,omitempty"`
}
```

## 2. Edge Configuration & Dynamic Navigation

### Edge Types

The system supports three types of edges for different navigation scenarios:

#### Always Accessible Edges (Independent Nodes)
```go
// Always accessible edges (independent nodes)
edges = append(edges, types.NewEdge(startNode.ID, panNode.ID, types.EdgeCondition{
    Type: "always", // No conditions
}))
```

#### Conditional Edges (Dependent Nodes)
```go
// Conditional edges (dependent nodes)
edges = append(edges, types.NewEdge(startNode.ID, businessDocumentNode.ID, types.EdgeCondition{
    Type:     "field_value",    // Condition type
    Field:    "business_type",  // Field to check
    Operator: "ne",            // Not equal
    Value:    "",              // Empty value (must be filled)
}))
```

#### Completion Edges
```go
// Completion edges (from any node to completion)
edges = append(edges, types.NewEdge(panNode.ID, completionNode.ID, types.EdgeCondition{
    Type: "completion_check", // Special type for completion validation
}))
```

### Edge Condition Types

```go
type EdgeCondition struct {
    Type     string `json:"type"`     // "always", "field_value", "completion_check"
    Field    string `json:"field"`    // Field to check (for field_value type)
    Operator string `json:"operator"` // "eq", "ne", "in", "not_in"
    Value    string `json:"value"`    // Value to compare against
}
```

### Edge Relationship Building

```go
// After creating edges, update node relationships
for _, edge := range edges {
    graph.Edges[edge.ID] = edge
    
    // Update node edge relationships
    fromNode := graph.Nodes[edge.FromNodeID]
    toNode := graph.Nodes[edge.ToNodeID]
    
    if fromNode != nil {
        fromNode.OutgoingEdges = append(fromNode.OutgoingEdges, edge.ID)
    }
    if toNode != nil {
        toNode.IncomingEdges = append(toNode.IncomingEdges, edge.ID)
    }
}
```

## 3. Node Validation Logic on Save

### Validation Flow in SubmitNodeData

The validation process follows a structured flow when a user submits node data:

```go
func (s *Service) SubmitNodeData(ctx context.Context, sessionID string, data map[string]interface{}) (*NextStepResult, error) {
    // 1. Get session and current node
    session, err := s.storage.GetSession(ctx, sessionID)
    graph, err := s.storage.GetGraph(ctx, session.GraphID)
    currentNode, exists := graph.Nodes[session.CurrentNodeID]
    
    // 2. Merge session data with new data for validation
    validationData := make(map[string]interface{})
    for k, v := range session.Data {
        validationData[k] = v
    }
    for k, v := range data {
        validationData[k] = v
    }
    
    // 3. Validate current node data
    validationResult := s.engine.ValidateNode(ctx, currentNode, validationData)
    if !validationResult.Valid {
        return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
    }
    
    // 4. Update session data
    for key, value := range data {
        session.Data[key] = value
    }
    
    // 5. Check path completeness and rule groups
    pathValid, missingNodes := s.engine.ValidatePathCompleteness(ctx, graph, session.CurrentNodeID, session.Data, session.History)
    
    // 6. Check rule groups for production onboarding
    if !pathValid && businessType, hasBusinessType := session.Data["business_type"]; hasBusinessType {
        ruleGroupPassed, _ := s.engine.ValidateProductionOnboardingCompleteness(ctx, graph, session.CurrentNodeID, session.Data, businessType)
        if ruleGroupPassed {
            session.Status = SessionStatusCompleted
        }
    }
}
```

### Node-Level Validation

```go
func (e *Engine) ValidateNode(ctx context.Context, node *Node, data map[string]interface{}) *ValidationResult {
    result := &ValidationResult{Valid: true, Errors: []ValidationError{}}
    
    // 1. Check required fields
    for _, fieldID := range node.Validation.RequiredFields {
        if value, exists := data[fieldID]; !exists || value == nil || value == "" {
            result.Errors = append(result.Errors, ValidationError{
                Field:   fieldID,
                Message: fmt.Sprintf("Field %s is required", fieldID),
            })
        }
    }
    
    // 2. Check conditional fields
    for _, condition := range node.Validation.Conditions {
        if e.evaluateCondition(condition, data) {
            // Condition met, field is required
            if value, exists := data[condition.Field]; !exists || value == nil || value == "" {
                result.Errors = append(result.Errors, ValidationError{
                    Field:   condition.Field,
                    Message: condition.Rule,
                })
            }
        }
    }
    
    return result
}
```

### Validation Result Structure

```go
type ValidationResult struct {
    Valid     bool              `json:"valid"`
    Errors    []ValidationError `json:"errors,omitempty"`
    Warnings  []ValidationWarning `json:"warnings,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code,omitempty"`
}
```

## 4. Observer Pattern for Dynamic Node Status

### Dynamic Node Structure

```go
type DynamicNode struct {
    *Node
    Status        NodeStatus       `json:"status"`        // Current status
    InitialStatus NodeStatus       `json:"initial_status"` // Initial status
    Dependencies  []NodeDependency `json:"dependencies"`  // What this node depends on
    Observers     []NodeObserver   `json:"-"`            // Observers for status changes
    mu            sync.RWMutex     `json:"-"`            // Thread safety
}

type NodeObserver interface {
    OnNodeStatusChanged(nodeID string, oldStatus, newStatus NodeStatus, sessionData map[string]interface{})
    OnNodeCompleted(nodeID string, sessionData map[string]interface{})
    OnNodeDataChanged(nodeID string, fieldID string, value interface{}, sessionData map[string]interface{})
}
```

### Node Status Types

```go
const (
    NodeStatusMandatory NodeStatus = "mandatory"  // Must be completed
    NodeStatusOptional  NodeStatus = "optional"   // Can be skipped
    NodeStatusDependent NodeStatus = "dependent"  // Depends on other conditions
    NodeStatusCompleted NodeStatus = "completed"  // Already completed
    NodeStatusDisabled  NodeStatus = "disabled"   // Not available
)
```

### Dynamic Status Evaluation

```go
func (de *DynamicEngine) EvaluateNodeDependencies(dynamicNode *DynamicNode, sessionData map[string]interface{}) NodeStatus {
    // Check all dependencies
    for _, dependency := range dynamicNode.Dependencies {
        if de.isDependencySatisfied(dependency, sessionData) {
            return NodeStatusMandatory // Dependency satisfied, node becomes mandatory
        }
    }
    
    // If no dependencies make it mandatory, it's optional or dependent
    if len(dynamicNode.Dependencies) > 0 {
        return NodeStatusDependent // Has dependencies but none satisfied
    }
    
    return NodeStatusOptional // No dependencies, optional
}
```

### Observer Notification

```go
func (dg *DynamicGraph) OnNodeDataChanged(nodeID string, fieldID string, value interface{}, sessionData map[string]interface{}) {
    // Notify all observers of the data change
    for _, observer := range dg.DynamicNodes[nodeID].Observers {
        observer.OnNodeDataChanged(nodeID, fieldID, value, sessionData)
    }
    
    // Re-evaluate dependent nodes
    for _, dynamicNode := range dg.DynamicNodes {
        if dynamicNode.Status != NodeStatusCompleted {
            newStatus := dg.dynamicEngine.EvaluateNodeDependencies(dynamicNode, sessionData)
            if newStatus != dynamicNode.Status {
                dg.UpdateNodeStatus(dynamicNode.ID, newStatus, sessionData)
            }
        }
    }
}
```

### Node Dependency Structure

```go
type NodeDependency struct {
    FieldID      string      `json:"field_id"`                // Field to check
    Operator     string      `json:"operator"`                // eq, ne, in, not_in, etc.
    Value        interface{} `json:"value"`                   // Value to compare against
    Condition    string      `json:"condition"`               // Human readable condition
    BusinessType string      `json:"business_type,omitempty"` // Business type specific dependency
}
```

## 5. Rule Group Validation for Business Type Completion

### Rule Group Structure

```go
type RuleGroup struct {
    ID                string                          `json:"id"`
    Name              string                          `json:"name"`
    Description       string                          `json:"description"`
    RequiredNodes     []string                        `json:"required_nodes"`     // Node IDs that must be completed
    RequiredFields    map[string][]string             `json:"required_fields"`    // Node ID -> []Field IDs
    ConditionalFields map[string]ConditionalFieldRule `json:"conditional_fields"` // Conditional requirements
}
```

### Business Type Rule Groups

```go
func (e *Engine) GetBusinessTypeRuleGroups(businessType string) []RuleGroup {
    allRuleGroups := e.getAllRuleGroups()
    baseRuleGroups := allRuleGroups["individual"] // Base template
    
    // Customize for specific business type
    if businessType == "proprietorship" {
        return []RuleGroup{
            {
                ID:          "proprietorship_basic_path",
                Name:        "Proprietorship Basic Path",
                Description: "Basic requirements for proprietorship onboarding",
                RequiredNodes: []string{"business_type_selection", "pan_number_node_id", "payment_channel_node_id", "business_info_node_id"},
                RequiredFields: map[string][]string{
                    "business_type_selection": {"business_type"},
                    "pan_number_node_id":      {"pan_number", "pan_document"},
                    "payment_channel_node_id": {"payment_channel"},
                    "business_info_node_id":   {"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
                },
            },
        }
    }
    
    return baseRuleGroups
}
```

### Rule Group Evaluation

```go
func (e *Engine) ValidateProductionOnboardingCompleteness(ctx context.Context, graph *Graph, currentNodeID string, sessionData map[string]interface{}, businessType interface{}) (bool, []string) {
    businessTypeStr := fmt.Sprintf("%v", businessType)
    ruleGroups := e.GetBusinessTypeRuleGroups(businessTypeStr)
    
    // Check if ANY rule group passes
    for _, ruleGroup := range ruleGroups {
        ruleGroupPassed, _ := e.EvaluateRuleGroup(ruleGroup, sessionData)
        if ruleGroupPassed {
            // This rule group passes - user is eligible for completion
            return true, []string{}
        }
    }
    
    return false, []string{"No rule group passed"}
}

func (e *Engine) EvaluateRuleGroup(ruleGroup RuleGroup, sessionData map[string]interface{}) (bool, []string) {
    missingRequirements := make([]string, 0)
    
    // Check required nodes and fields
    for _, nodeID := range ruleGroup.RequiredNodes {
        requiredFields, hasRequiredFields := ruleGroup.RequiredFields[nodeID]
        if !hasRequiredFields {
            continue
        }
        
        for _, fieldID := range requiredFields {
            if value, exists := sessionData[fieldID]; !exists || value == nil || value == "" {
                missingRequirements = append(missingRequirements, fmt.Sprintf("Node %s: field %s is required", nodeID, fieldID))
            }
        }
    }
    
    return len(missingRequirements) == 0, missingRequirements
}
```

### Conditional Field Rules

```go
type ConditionalFieldRule struct {
    NodeID      string `json:"node_id"`
    FieldID     string `json:"field_id"`
    Condition   string `json:"condition"` // Field to check
    Operator    string `json:"operator"`  // eq, ne, etc.
    Value       string `json:"value"`     // Value to compare against
    Description string `json:"description"`
}
```

## 6. Cross-Node Validation Integration

### Cross-Node Validation Rules

```go
type CrossNodeValidationRule struct {
    ID          string                    `json:"id"`
    Name        string                    `json:"name"`
    Description string                    `json:"description"`
    Fields      []CrossNodeFieldReference `json:"fields"`                // Fields from different nodes
    Condition   CrossNodeCondition        `json:"condition"`             // Validation logic
    ErrorMsg    string                    `json:"error_msg"`             // Error message
    BusinessType string                   `json:"business_type,omitempty"` // Business type specific
    Severity    ValidationSeverity        `json:"severity"`              // Error, Warning, Info
    Enabled     bool                      `json:"enabled"`               // Active status
}
```

### Cross-Node Field Reference

```go
type CrossNodeFieldReference struct {
    NodeID  string `json:"node_id"`  // ID of the node containing the field
    FieldID string `json:"field_id"` // ID of the field within the node
    Alias   string `json:"alias"`    // Optional alias for the field in the condition
}
```

### Cross-Node Condition

```go
type CrossNodeCondition struct {
    Type     string                 `json:"type"`     // Condition type: "field_match", "field_contains", "custom_logic"
    Operator string                 `json:"operator"` // Comparison operator: "eq", "ne", "contains", "matches", "custom"
    Fields   []string               `json:"fields"`   // Field aliases to compare (from CrossNodeFieldReference.Alias)
    Value    interface{}            `json:"value,omitempty"` // Static value to compare against (if applicable)
    Logic    string                 `json:"logic,omitempty"` // Custom validation logic (for complex cases)
    Metadata map[string]interface{} `json:"metadata,omitempty"` // Additional metadata for the condition
}
```

### Validation Severity Levels

```go
type ValidationSeverity string

const (
    ValidationSeverityError   ValidationSeverity = "error"   // Blocks completion
    ValidationSeverityWarning ValidationSeverity = "warning" // Shows warning but allows continuation
    ValidationSeverityInfo    ValidationSeverity = "info"    // Informational only
)
```

### Cross-Node Validation Integration

```go
func (cve *CrossNodeValidationEngine) ValidateCrossNodeRules(ctx context.Context, graph *types.Graph, sessionData map[string]interface{}, businessType string) ([]CrossNodeValidationResult, error) {
    var results []CrossNodeValidationResult
    
    for _, rule := range graph.CrossNodeValidation {
        if !rule.Enabled {
            continue
        }
        
        // Check if rule applies to current business type
        if rule.BusinessType != "" && rule.BusinessType != businessType {
            continue
        }
        
        // Only validate if all required fields are present in session data
        if cve.allRequiredFieldsPresent(rule, sessionData) {
            result := cve.validateSingleRule(ctx, rule, sessionData, graph)
            results = append(results, result)
        }
    }
    
    return results, nil
}
```

### Example Cross-Node Validation Rules

#### Business Name ↔ Bank Name Consistency
```go
{
    ID:          "business_name_bank_name_match",
    Name:        "Business Name and Bank Name Consistency",
    Fields: []CrossNodeFieldReference{
        {
            NodeID:  "business_info_node_id",
            FieldID: "business_name",
            Alias:   "business_name",
        },
        {
            NodeID:  "bank_account_node_id", 
            FieldID: "bank_name",
            Alias:   "bank_name",
        },
    },
    Condition: CrossNodeCondition{
        Type:     "custom_logic",
        Logic:    "business_name_matches_bank_name",
    },
    BusinessType: "", // Applies to all business types
    Severity:     ValidationSeverityWarning,
    Enabled:      true,
}
```

#### PAN Consistency (Individual only)
```go
{
    ID:          "pan_signatory_consistency",
    Name:        "PAN and Signatory PAN Consistency",
    BusinessType: "individual", // Only for individual business type
    Fields: []CrossNodeFieldReference{
        {
            NodeID:  "pan_number_node_id",
            FieldID: "pan_number",
            Alias:   "pan_number",
        },
        {
            NodeID:  "authorised_signatory_node_id",
            FieldID: "signatory_pan", 
            Alias:   "signatory_pan",
        },
    },
    Condition: CrossNodeCondition{
        Type:     "custom_logic",
        Logic:    "pan_matches_signatory_pan",
    },
    Severity: ValidationSeverityError,
    Enabled:  true,
}
```

## Complete Validation Flow

The complete validation process follows this structured flow:

1. **User submits node data** → `SubmitNodeData()`
2. **Node-level validation** → `ValidateNode()` (required fields, conditional fields)
3. **Update session data** → Merge new data with existing session data
4. **Path completeness check** → `ValidatePathCompleteness()`
5. **Rule group validation** → `ValidateProductionOnboardingCompleteness()`
6. **Cross-node validation** → `ValidateCrossNodeRules()` (if applicable)
7. **Observer notifications** → Update dependent node statuses
8. **Navigation decision** → Move to next node or mark complete

### Validation Flow Diagram

```
User Input → Node Validation → Session Update → Path Check → Rule Groups → Cross-Node → Observer → Navigation
     ↓              ↓              ↓            ↓           ↓           ↓         ↓         ↓
  Submit Data → Validate Fields → Save Data → Check Path → Check Rules → Validate → Notify → Next/Complete
```

## Key Benefits

- ✅ **Flexible Navigation** - Users can fill forms in any order
- ✅ **Dynamic Dependencies** - Nodes become mandatory/optional based on user actions
- ✅ **Business Type Specific** - Different rules for different business types
- ✅ **Cross-Node Validation** - Data consistency across multiple nodes
- ✅ **Rule Group System** - Multiple paths to completion
- ✅ **Observer Pattern** - Real-time status updates
- ✅ **Persistence** - State maintained across sessions
- ✅ **Extensible** - Easy to add new validation rules and business types

## API Endpoints

### Standard Onboarding API

```bash
# Start a session
POST /api/v1/sessions
{
  "graph_id": "graph-id",
  "user_id": "user-id"
}

# Submit node data
POST /api/v1/sessions/{session_id}/submit
{
  "field1": "value1",
  "field2": "value2"
}

# Get current node
GET /api/v1/sessions/{session_id}/current

# Get eligible nodes
GET /api/v1/sessions/{session_id}/eligible-nodes
```

### Dynamic Onboarding API

```bash
# Start dynamic session
POST /api/v1/dynamic/sessions
{
  "graph_id": "graph-id",
  "user_id": "user-id"
}

# Submit dynamic node data
POST /api/v1/dynamic/sessions/{session_id}/submit
{
  "field1": "value1",
  "field2": "value2"
}

# Get dynamic node status
GET /api/v1/dynamic/sessions/{session_id}/status

# Update business type
PUT /api/v1/dynamic/sessions/{session_id}/business-type
{
  "business_type": "proprietorship"
}

# Get dynamic state summary
GET /api/v1/dynamic/sessions/{session_id}/summary
```

## Testing

### Test Files

- `test-dynamic-navigation.html` - Test dynamic navigation and form rendering
- `test-edge-visualization.html` - Visualize graph structure and edge relationships
- `test-cross-node-validation.html` - Test cross-node validation rules

### Running Tests

```bash
# Start the backend
go run main.go

# Start the frontend server
python3 -m http.server 3000

# Open test pages
open http://localhost:3000/test-dynamic-navigation.html
open http://localhost:3000/test-edge-visualization.html
open http://localhost:3000/test-cross-node-validation.html
```

### Test Scenarios

1. **Dynamic Navigation Test**
   - Start session with Enhanced Dynamic Production Onboarding graph
   - Fill forms in different orders
   - Verify node status changes
   - Test completion with different business types

2. **Edge Visualization Test**
   - View graph structure
   - Analyze node categories
   - Check edge relationships
   - Verify node dependencies

3. **Cross-Node Validation Test**
   - Fill business information and bank details
   - Test name consistency validation
   - Verify business type specific rules
   - Check validation severity levels

## Architecture Diagrams

### System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend UI   │    │   API Layer     │    │  Service Layer  │
│                 │    │                 │    │                 │
│ - HTML/JS       │◄──►│ - REST API      │◄──►│ - Service       │
│ - Form Builder  │    │ - Validation    │    │ - Engine        │
│ - Navigation    │    │ - Error Handling│    │ - Dynamic Engine│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Storage Layer  │
                       │                 │
                       │ - Sessions      │
                       │ - Graphs        │
                       │ - Persistence   │
                       └─────────────────┘
```

### Validation Flow Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Node Level    │    │  Rule Group     │    │ Cross-Node      │
│   Validation    │    │  Validation     │    │ Validation      │
│                 │    │                 │    │                 │
│ - Required      │    │ - Business Type │    │ - Field Match   │
│ - Conditional   │    │ - Required Nodes│    │ - Custom Logic  │
│ - Field Rules   │    │ - Field Rules   │    │ - Severity      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                       ┌─────────────────┐
                       │  Observer       │
                       │  Pattern        │
                       │                 │
                       │ - Status Update │
                       │ - Dependency    │
                       │ - Notification  │
                       └─────────────────┘
```

### Node Status Flow

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Initial       │    │   Dynamic       │    │   Final         │
│   Status        │    │   Status        │    │   Status        │
│                 │    │                 │    │                 │
│ - Mandatory     │    │ - Re-evaluate   │    │ - Completed     │
│ - Optional      │    │ - Dependencies  │    │ - Disabled      │
│ - Dependent     │    │ - Observer      │    │ - Skipped       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                       ┌─────────────────┐
                       │  User Actions   │
                       │                 │
                       │ - Fill Forms    │
                       │ - Submit Data   │
                       │ - Change Values │
                       └─────────────────┘
```

## Getting Started

### Prerequisites

- Go 1.19+
- Python 3.7+ (for frontend server)
- Modern web browser

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd onboarding_grabh_based

# Install dependencies
go mod tidy

# Start the backend
go run main.go

# In another terminal, start the frontend server
python3 -m http.server 3000
```

### Quick Start

1. **Start the system**:
   ```bash
   go run main.go
   python3 -m http.server 3000
   ```

2. **Open test interface**:
   ```bash
   open http://localhost:3000/test-dynamic-navigation.html
   ```

3. **Test the flow**:
   - Select "Enhanced Dynamic Production Onboarding" graph
   - Start a session
   - Fill forms in any order
   - Observe dynamic status changes
   - Test completion with different business types

### Configuration

The system can be configured through:

- **Graph Definition**: `examples/enhanced_dynamic_production_onboarding.go`
- **Validation Rules**: `config/validation_rules.yaml`
- **Business Logic**: `internal/onboarding/engine.go`
- **Cross-Node Rules**: `examples/enhanced_dynamic_production_onboarding.go`

## Contributing

### Adding New Business Types

1. **Update node definitions** in `enhanced_dynamic_production_onboarding.go`
2. **Add rule groups** in `engine.go`
3. **Update cross-node validation** rules
4. **Test with new business type**

### Adding New Validation Rules

1. **Define rule structure** in `types.go`
2. **Implement validation logic** in `cross_node_validation.go`
3. **Add to graph definition**
4. **Test validation scenarios**

### Adding New Node Types

1. **Update node type constants** in `types.go`
2. **Add node creation logic** in graph builders
3. **Update validation logic** if needed
4. **Test node behavior**

## Troubleshooting

### Common Issues

1. **Session not found**: Check if session ID is correct and session exists
2. **Validation errors**: Verify field names and validation rules
3. **Navigation issues**: Check edge conditions and node dependencies
4. **Cross-node validation fails**: Ensure all required fields are present

### Debug Mode

Enable debug logging by setting log level to debug:

```go
logger.SetLevel(logrus.DebugLevel)
```

### Log Analysis

Key log patterns to look for:

- `Node validation failed` - Field validation issues
- `Rule group validation failed` - Business type requirements not met
- `Cross-node validation rule completed` - Cross-node validation results
- `Dynamic state saved to session` - Persistence operations

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support:

1. Check the troubleshooting section
2. Review the test files for examples
3. Examine the debug logs
4. Create an issue with detailed information

---

**Happy Coding! 🚀**
