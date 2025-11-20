package onboarding

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// NodeStatus represents the current status of a node
type NodeStatus string

const (
	NodeStatusMandatory NodeStatus = "mandatory"
	NodeStatusOptional  NodeStatus = "optional"
	NodeStatusDependent NodeStatus = "dependent"
	NodeStatusCompleted NodeStatus = "completed"
	NodeStatusDisabled  NodeStatus = "disabled"
)

// NodeDependency represents a dependency condition for a node
type NodeDependency struct {
	FieldID      string      `json:"field_id"`                // Field to check
	Operator     string      `json:"operator"`                // eq, ne, in, not_in, etc.
	Value        interface{} `json:"value"`                   // Value to compare against
	Condition    string      `json:"condition"`               // Human readable condition
	BusinessType string      `json:"business_type,omitempty"` // Business type specific dependency
}

// DynamicNode represents a node with dynamic status tracking
type DynamicNode struct {
	*Node
	Status        NodeStatus       `json:"status"`
	InitialStatus NodeStatus       `json:"initial_status"`
	Dependencies  []NodeDependency `json:"dependencies"`
	Observers     []NodeObserver   `json:"-"` // Not serialized
	mu            sync.RWMutex     `json:"-"` // Not serialized
}

// NodeObserver interface for observing node status changes
type NodeObserver interface {
	OnNodeStatusChanged(nodeID string, oldStatus, newStatus NodeStatus, sessionData map[string]interface{})
	OnNodeCompleted(nodeID string, sessionData map[string]interface{})
	OnNodeDataChanged(nodeID string, fieldID string, value interface{}, sessionData map[string]interface{})
}

// DynamicGraph represents a graph with dynamic node management
type DynamicGraph struct {
	*Graph
	DynamicNodes map[string]*DynamicNode `json:"dynamic_nodes"`
	Observers    []NodeObserver          `json:"-"` // Not serialized
	mu           sync.RWMutex            `json:"-"` // Not serialized
}

// DynamicEngine manages dynamic node status and validation
type DynamicEngine struct {
	*Engine
	logger *logrus.Logger
}

// NewDynamicEngine creates a new dynamic engine
func NewDynamicEngine(logger *logrus.Logger) *DynamicEngine {
	return &DynamicEngine{
		Engine: NewEngine(logger),
		logger: logger,
	}
}

// ConvertToDynamicGraph converts a regular graph to a dynamic graph
func (de *DynamicEngine) ConvertToDynamicGraph(graph *Graph, businessType string) *DynamicGraph {
	dynamicGraph := &DynamicGraph{
		Graph:        graph,
		DynamicNodes: make(map[string]*DynamicNode),
		Observers:    make([]NodeObserver, 0),
	}

	// Convert each node to a dynamic node
	for nodeID, node := range graph.Nodes {
		dynamicNode := &DynamicNode{
			Node:          node,
			Status:        de.determineInitialStatus(node, businessType),
			InitialStatus: de.determineInitialStatus(node, businessType),
			Dependencies:  de.extractDependencies(node, businessType),
			Observers:     make([]NodeObserver, 0),
		}
		dynamicGraph.DynamicNodes[nodeID] = dynamicNode
	}

	// Add the graph as an observer to all nodes
	dynamicGraph.AddObserver(dynamicGraph)

	return dynamicGraph
}

// determineInitialStatus determines the initial status of a node based on business type
func (de *DynamicEngine) determineInitialStatus(node *Node, businessType string) NodeStatus {
	// Check if this is a start node
	if node.Type == "start" {
		return NodeStatusMandatory
	}

	// Check if this is an end node
	if node.Type == "end" {
		return NodeStatusOptional // End nodes are optional until all required nodes are completed
	}

	// Check if node has business type specific requirements
	if de.isNodeRequiredForBusinessType(node, businessType) {
		return NodeStatusMandatory
	}

	// Check if node has dependencies
	if de.hasDependencies(node) {
		return NodeStatusDependent
	}

	// Default to optional
	return NodeStatusOptional
}

// isNodeRequiredForBusinessType checks if a node is required for a specific business type
func (de *DynamicEngine) isNodeRequiredForBusinessType(node *Node, businessType string) bool {
	// Define business type specific requirements
	businessTypeRequirements := map[string][]string{
		"individual": {
			"PAN Number",
			"Payment Channel",
			"Business Information",
		},
		"proprietorship": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Information",
			"Bank Account Details",
		},
		"private_limited": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Document",
			"Authorised Signatory Details",
			"Bank Account Details",
			"Business Information",
		},
		"public_limited": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Document",
			"Authorised Signatory Details",
			"Bank Account Details",
			"Business Information",
		},
		"partnership": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Document",
			"Authorised Signatory Details",
			"Bank Account Details",
			"Business Information",
		},
		"llp": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Document",
			"Authorised Signatory Details",
			"Bank Account Details",
			"Business Information",
		},
		"trust": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Document",
			"Authorised Signatory Details",
			"Bank Account Details",
			"Business Information",
		},
		"society": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Document",
			"Authorised Signatory Details",
			"Bank Account Details",
			"Business Information",
		},
		"huf": {
			"PAN Number",
			"Payment Channel",
			"MCC & Policy Verification",
			"Business Document",
			"Authorised Signatory Details",
			"Bank Account Details",
			"Business Information",
		},
	}

	requiredNodes, exists := businessTypeRequirements[businessType]
	if !exists {
		return false
	}

	for _, requiredNode := range requiredNodes {
		if node.Name == requiredNode {
			return true
		}
	}

	return false
}

// hasDependencies checks if a node has dependencies
func (de *DynamicEngine) hasDependencies(node *Node) bool {
	// Check if node has validation conditions that create dependencies
	if len(node.Validation.Conditions) > 0 {
		return true
	}

	// Check if node has conditional fields
	for _, field := range node.Fields {
		if field.Required && len(field.Validation.CustomRules) > 0 {
			return true
		}
	}

	return false
}

// extractDependencies extracts dependencies from a node
func (de *DynamicEngine) extractDependencies(node *Node, businessType string) []NodeDependency {
	dependencies := make([]NodeDependency, 0)

	// Extract from validation conditions
	for _, condition := range node.Validation.Conditions {
		dependency := NodeDependency{
			FieldID:      condition.Field,
			Operator:     condition.Operator,
			Value:        condition.Value,
			Condition:    condition.Rule,
			BusinessType: businessType,
		}
		dependencies = append(dependencies, dependency)
	}

	// Extract from field conditions
	for _, field := range node.Fields {
		if field.Required && len(field.Validation.CustomRules) > 0 {
			// Add field-specific dependencies
			for _, rule := range field.Validation.CustomRules {
				dependency := NodeDependency{
					FieldID:      field.ID,
					Operator:     "custom",
					Value:        rule,
					Condition:    fmt.Sprintf("Field %s requires %s", field.ID, rule),
					BusinessType: businessType,
				}
				dependencies = append(dependencies, dependency)
			}
		}
	}

	return dependencies
}

// AddObserver adds an observer to the dynamic graph
func (dg *DynamicGraph) AddObserver(observer NodeObserver) {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	dg.Observers = append(dg.Observers, observer)
}

// RemoveObserver removes an observer from the dynamic graph
func (dg *DynamicGraph) RemoveObserver(observer NodeObserver) {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	for i, obs := range dg.Observers {
		if obs == observer {
			dg.Observers = append(dg.Observers[:i], dg.Observers[i+1:]...)
			break
		}
	}
}

// UpdateNodeStatus updates the status of a node and notifies observers
func (dg *DynamicGraph) UpdateNodeStatus(nodeID string, newStatus NodeStatus, sessionData map[string]interface{}) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	dynamicNode, exists := dg.DynamicNodes[nodeID]
	if !exists {
		return
	}

	dynamicNode.mu.Lock()
	oldStatus := dynamicNode.Status
	dynamicNode.Status = newStatus
	dynamicNode.mu.Unlock()

	// Notify observers
	for _, observer := range dg.Observers {
		observer.OnNodeStatusChanged(nodeID, oldStatus, newStatus, sessionData)
	}
}

// OnNodeStatusChanged implements NodeObserver interface for the graph itself
func (dg *DynamicGraph) OnNodeStatusChanged(nodeID string, oldStatus, newStatus NodeStatus, sessionData map[string]interface{}) {
	// Don't recursively evaluate dependencies to avoid deadlock
	// Dependencies will be evaluated when the session is loaded or when explicitly requested
}

// OnNodeCompleted implements NodeObserver interface
func (dg *DynamicGraph) OnNodeCompleted(nodeID string, sessionData map[string]interface{}) {
	// Mark node as completed
	dg.UpdateNodeStatus(nodeID, NodeStatusCompleted, sessionData)

	// Evaluate dependent nodes
	de := &DynamicEngine{}
	de.evaluateDependentNodes(dg, sessionData)
}

// OnNodeDataChanged implements NodeObserver interface
func (dg *DynamicGraph) OnNodeDataChanged(nodeID string, fieldID string, value interface{}, sessionData map[string]interface{}) {
	// Evaluate dependent nodes when data changes
	de := &DynamicEngine{}
	de.evaluateDependentNodes(dg, sessionData)
}

// evaluateDependentNodes evaluates and updates the status of dependent nodes
func (de *DynamicEngine) evaluateDependentNodes(dg *DynamicGraph, sessionData map[string]interface{}) {
	for nodeID, dynamicNode := range dg.DynamicNodes {
		if dynamicNode.Status == NodeStatusDependent {
			newStatus := de.EvaluateNodeDependencies(dynamicNode, sessionData)
			if newStatus != dynamicNode.Status {
				dg.UpdateNodeStatus(nodeID, newStatus, sessionData)
			}
		}
	}
}

// EvaluateNodeDependencies evaluates if a dependent node should become mandatory or optional
func (de *DynamicEngine) EvaluateNodeDependencies(dynamicNode *DynamicNode, sessionData map[string]interface{}) NodeStatus {
	// Check all dependencies
	for _, dependency := range dynamicNode.Dependencies {
		if de.isDependencySatisfied(dependency, sessionData) {
			return NodeStatusMandatory
		}
	}

	// If no dependencies are satisfied, check if node should be optional
	if de.shouldNodeBeOptional(dynamicNode, sessionData) {
		return NodeStatusOptional
	}

	return NodeStatusDependent
}

// isDependencySatisfied checks if a dependency condition is satisfied
func (de *DynamicEngine) isDependencySatisfied(dependency NodeDependency, sessionData map[string]interface{}) bool {
	fieldValue, exists := sessionData[dependency.FieldID]
	if !exists {
		return false
	}

	switch dependency.Operator {
	case "eq":
		return fieldValue == dependency.Value
	case "ne":
		return fieldValue != dependency.Value
	case "in":
		if values, ok := dependency.Value.([]interface{}); ok {
			for _, v := range values {
				if fieldValue == v {
					return true
				}
			}
		}
		return false
	case "not_in":
		if values, ok := dependency.Value.([]interface{}); ok {
			for _, v := range values {
				if fieldValue == v {
					return false
				}
			}
		}
		return true
	case "custom":
		// Handle custom rules - for now, return true if the field exists
		_, exists := sessionData[dependency.FieldID]
		return exists
	default:
		return false
	}
}

// shouldNodeBeOptional determines if a node should be optional based on current state
func (de *DynamicEngine) shouldNodeBeOptional(dynamicNode *DynamicNode, sessionData map[string]interface{}) bool {
	// For now, return false - this can be enhanced with more complex logic
	// In a real implementation, this would check if all mandatory dependencies are satisfied
	return false
}

// GetCompletionStatus returns the current completion status of the dynamic graph
func (dg *DynamicGraph) GetCompletionStatus() map[string]interface{} {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	status := map[string]interface{}{
		"total_nodes":     len(dg.DynamicNodes),
		"mandatory_nodes": 0,
		"optional_nodes":  0,
		"dependent_nodes": 0,
		"completed_nodes": 0,
		"disabled_nodes":  0,
		"can_complete":    false,
	}

	mandatoryCompleted := 0
	totalMandatory := 0

	for _, dynamicNode := range dg.DynamicNodes {
		switch dynamicNode.Status {
		case NodeStatusMandatory:
			status["mandatory_nodes"] = status["mandatory_nodes"].(int) + 1
			totalMandatory++
			if dynamicNode.Status == NodeStatusCompleted {
				mandatoryCompleted++
			}
		case NodeStatusOptional:
			status["optional_nodes"] = status["optional_nodes"].(int) + 1
		case NodeStatusDependent:
			status["dependent_nodes"] = status["dependent_nodes"].(int) + 1
		case NodeStatusCompleted:
			status["completed_nodes"] = status["completed_nodes"].(int) + 1
		case NodeStatusDisabled:
			status["disabled_nodes"] = status["disabled_nodes"].(int) + 1
		}
	}

	// Can complete if all mandatory nodes are completed
	status["can_complete"] = mandatoryCompleted == totalMandatory && totalMandatory > 0

	return status
}
