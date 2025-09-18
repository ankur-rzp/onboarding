package onboarding

import (
	"context"
	"fmt"
	"strings"

	"onboarding-system/internal/types"

	"github.com/sirupsen/logrus"
)

// AdvancedEngine handles the sophisticated graph traversal with multiple entry points and activation rules
type AdvancedEngine struct {
	logger *logrus.Logger
}

// NewAdvancedEngine creates a new advanced graph engine
func NewAdvancedEngine(logger *logrus.Logger) *AdvancedEngine {
	return &AdvancedEngine{
		logger: logger,
	}
}

// ActivationRule represents a rule that determines if a user can be activated
type ActivationRule struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Conditions    []ActivationCondition  `json:"conditions"`
	RequiredNodes []string               `json:"required_nodes"` // Nodes that must be visited
	ExcludedNodes []string               `json:"excluded_nodes"` // Nodes that should not be visited
	Metadata      map[string]interface{} `json:"metadata"`
}

// ActivationCondition represents a condition for activation
type ActivationCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Required bool        `json:"required"` // If true, this condition must be met
}

// NodeRule represents rules that apply when a node is visited
type NodeRule struct {
	ID         string                 `json:"id"`
	NodeID     string                 `json:"node_id"`
	RuleType   string                 `json:"rule_type"` // "validation", "path_limitation", "activation_check"
	Conditions []ActivationCondition  `json:"conditions"`
	Actions    []RuleAction           `json:"actions"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// RuleAction represents an action to take when a rule is triggered
type RuleAction struct {
	Type     string                 `json:"type"`   // "disable_edge", "require_node", "exclude_node", "check_activation"
	Target   string                 `json:"target"` // Edge ID, Node ID, or Activation Rule ID
	Metadata map[string]interface{} `json:"metadata"`
}

// AdvancedGraph extends the basic graph with activation rules and node rules
type AdvancedGraph struct {
	*types.Graph
	ActivationRules []ActivationRule `json:"activation_rules"`
	NodeRules       []NodeRule       `json:"node_rules"`
	EntryNodes      []string         `json:"entry_nodes"` // Nodes that can be starting points
}

// AdvancedSession extends the basic session with activation tracking
type AdvancedSession struct {
	*types.Session
	VisitedNodes     []string   `json:"visited_nodes"`
	DisabledEdges    []string   `json:"disabled_edges"`
	ExcludedNodes    []string   `json:"excluded_nodes"`
	ActivationStatus string     `json:"activation_status"` // "pending", "activated", "failed"
	ActivationRules  []string   `json:"activation_rules"`  // Rules that have been checked
	PathHistory      []PathStep `json:"path_history"`
}

// PathStep represents a step in the user's path
type PathStep struct {
	NodeID        string                 `json:"node_id"`
	Timestamp     int64                  `json:"timestamp"`
	Data          map[string]interface{} `json:"data"`
	RulesApplied  []string               `json:"rules_applied"`
	EdgesDisabled []string               `json:"edges_disabled"`
}

// EvaluateNodeRules evaluates all rules for a specific node when it's visited
func (e *AdvancedEngine) EvaluateNodeRules(ctx context.Context, graph *AdvancedGraph, nodeID string, sessionData map[string]interface{}) (*RuleEvaluationResult, error) {
	result := &RuleEvaluationResult{
		Valid:            true,
		DisabledEdges:    make([]string, 0),
		RequiredNodes:    make([]string, 0),
		ExcludedNodes:    make([]string, 0),
		ActivationChecks: make([]string, 0),
		Metadata:         make(map[string]interface{}),
	}

	// Find rules for this node
	for _, rule := range graph.NodeRules {
		if rule.NodeID == nodeID {
			ruleResult, err := e.evaluateRule(ctx, rule, sessionData)
			if err != nil {
				e.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to evaluate rule")
				continue
			}

			// Apply rule actions
			for _, action := range ruleResult.TriggeredActions {
				switch action.Type {
				case "disable_edge":
					result.DisabledEdges = append(result.DisabledEdges, action.Target)
				case "require_node":
					result.RequiredNodes = append(result.RequiredNodes, action.Target)
				case "exclude_node":
					result.ExcludedNodes = append(result.ExcludedNodes, action.Target)
				case "check_activation":
					result.ActivationChecks = append(result.ActivationChecks, action.Target)
				}
			}
		}
	}

	return result, nil
}

// CheckActivation checks if the user can be activated based on current data and visited nodes
func (e *AdvancedEngine) CheckActivation(ctx context.Context, graph *AdvancedGraph, session *AdvancedSession) (*ActivationResult, error) {
	result := &ActivationResult{
		CanActivate:         false,
		ActivatedRules:      make([]string, 0),
		FailedRules:         make([]string, 0),
		MissingRequirements: make([]string, 0),
		Metadata:            make(map[string]interface{}),
	}

	// Check each activation rule
	for _, rule := range graph.ActivationRules {
		ruleResult, err := e.evaluateActivationRule(ctx, rule, session)
		if err != nil {
			e.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to evaluate activation rule")
			continue
		}

		if ruleResult.CanActivate {
			result.CanActivate = true
			result.ActivatedRules = append(result.ActivatedRules, rule.ID)
		} else {
			result.FailedRules = append(result.FailedRules, rule.ID)
			result.MissingRequirements = append(result.MissingRequirements, ruleResult.MissingRequirements...)
		}
	}

	return result, nil
}

// GetAvailablePaths returns the available paths from the current node based on rules
func (e *AdvancedEngine) GetAvailablePaths(ctx context.Context, graph *AdvancedGraph, currentNodeID string, session *AdvancedSession) ([]*types.Node, error) {
	availableNodes := make([]*types.Node, 0)

	// Get all edges from current node
	for _, edge := range graph.Edges {
		if edge.FromNodeID == currentNodeID {
			// Check if edge is disabled
			if e.isEdgeDisabled(edge.ID, session.DisabledEdges) {
				continue
			}

			// Check if target node is excluded
			if e.isNodeExcluded(edge.ToNodeID, session.ExcludedNodes) {
				continue
			}

			// Evaluate edge condition
			if e.evaluateEdgeCondition(edge.Condition, session.Data) {
				if targetNode, exists := graph.Nodes[edge.ToNodeID]; exists {
					availableNodes = append(availableNodes, targetNode)
				}
			}
		}
	}

	return availableNodes, nil
}

// GetEntryNodes returns all possible entry nodes for the graph
func (e *AdvancedEngine) GetEntryNodes(ctx context.Context, graph *AdvancedGraph) []*types.Node {
	entryNodes := make([]*types.Node, 0)

	for _, nodeID := range graph.EntryNodes {
		if node, exists := graph.Nodes[nodeID]; exists {
			entryNodes = append(entryNodes, node)
		}
	}

	return entryNodes
}

// ProcessNodeVisit processes a node visit and applies all relevant rules
func (e *AdvancedEngine) ProcessNodeVisit(ctx context.Context, graph *AdvancedGraph, session *AdvancedSession, nodeID string, data map[string]interface{}) error {
	// Add node to visited nodes if not already visited
	if !e.isNodeVisited(nodeID, session.VisitedNodes) {
		session.VisitedNodes = append(session.VisitedNodes, nodeID)
	}

	// Update session data
	for key, value := range data {
		session.Data[key] = value
	}

	// Evaluate node rules
	ruleResult, err := e.EvaluateNodeRules(ctx, graph, nodeID, session.Data)
	if err != nil {
		return fmt.Errorf("failed to evaluate node rules: %w", err)
	}

	// Apply rule results to session
	session.DisabledEdges = append(session.DisabledEdges, ruleResult.DisabledEdges...)
	session.ExcludedNodes = append(session.ExcludedNodes, ruleResult.ExcludedNodes...)

	// Add to path history
	pathStep := PathStep{
		NodeID:        nodeID,
		Timestamp:     e.getCurrentTimestamp(),
		Data:          data,
		RulesApplied:  make([]string, 0), // TODO: Track which rules were applied
		EdgesDisabled: ruleResult.DisabledEdges,
	}
	session.PathHistory = append(session.PathHistory, pathStep)

	// Check activation
	activationResult, err := e.CheckActivation(ctx, graph, session)
	if err != nil {
		return fmt.Errorf("failed to check activation: %w", err)
	}

	if activationResult.CanActivate {
		session.ActivationStatus = "activated"
		session.Status = types.SessionStatusCompleted
	}

	return nil
}

// Helper methods

type RuleEvaluationResult struct {
	Valid            bool                   `json:"valid"`
	DisabledEdges    []string               `json:"disabled_edges"`
	RequiredNodes    []string               `json:"required_nodes"`
	ExcludedNodes    []string               `json:"excluded_nodes"`
	ActivationChecks []string               `json:"activation_checks"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type ActivationResult struct {
	CanActivate         bool                   `json:"can_activate"`
	ActivatedRules      []string               `json:"activated_rules"`
	FailedRules         []string               `json:"failed_rules"`
	MissingRequirements []string               `json:"missing_requirements"`
	Metadata            map[string]interface{} `json:"metadata"`
}

type RuleResult struct {
	TriggeredActions []RuleAction           `json:"triggered_actions"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type ActivationRuleResult struct {
	CanActivate         bool                   `json:"can_activate"`
	MissingRequirements []string               `json:"missing_requirements"`
	Metadata            map[string]interface{} `json:"metadata"`
}

func (e *AdvancedEngine) evaluateRule(ctx context.Context, rule NodeRule, data map[string]interface{}) (*RuleResult, error) {
	result := &RuleResult{
		TriggeredActions: make([]RuleAction, 0),
		Metadata:         make(map[string]interface{}),
	}

	// Evaluate all conditions
	allConditionsMet := true
	for _, condition := range rule.Conditions {
		if !e.evaluateCondition(condition, data) {
			allConditionsMet = false
			if condition.Required {
				return result, nil // Required condition not met, rule doesn't apply
			}
		}
	}

	// If all conditions are met, trigger actions
	if allConditionsMet {
		result.TriggeredActions = rule.Actions
	}

	return result, nil
}

func (e *AdvancedEngine) evaluateActivationRule(ctx context.Context, rule ActivationRule, session *AdvancedSession) (*ActivationRuleResult, error) {
	result := &ActivationRuleResult{
		CanActivate:         false,
		MissingRequirements: make([]string, 0),
		Metadata:            make(map[string]interface{}),
	}

	// Check if all required nodes have been visited
	for _, requiredNode := range rule.RequiredNodes {
		if !e.isNodeVisited(requiredNode, session.VisitedNodes) {
			result.MissingRequirements = append(result.MissingRequirements, fmt.Sprintf("Required node not visited: %s", requiredNode))
		}
	}

	// Check if any excluded nodes have been visited
	for _, excludedNode := range rule.ExcludedNodes {
		if e.isNodeVisited(excludedNode, session.VisitedNodes) {
			result.MissingRequirements = append(result.MissingRequirements, fmt.Sprintf("Excluded node was visited: %s", excludedNode))
		}
	}

	// Evaluate all conditions
	allConditionsMet := true
	for _, condition := range rule.Conditions {
		if !e.evaluateCondition(condition, session.Data) {
			allConditionsMet = false
			if condition.Required {
				result.MissingRequirements = append(result.MissingRequirements, fmt.Sprintf("Required condition not met: %s %s %v", condition.Field, condition.Operator, condition.Value))
			}
		}
	}

	// If all requirements are met, user can be activated
	if len(result.MissingRequirements) == 0 && allConditionsMet {
		result.CanActivate = true
	}

	return result, nil
}

func (e *AdvancedEngine) evaluateCondition(condition ActivationCondition, data map[string]interface{}) bool {
	fieldValue, exists := data[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "eq":
		return fieldValue == condition.Value
	case "ne":
		return fieldValue != condition.Value
	case "in":
		if slice, ok := condition.Value.([]interface{}); ok {
			for _, v := range slice {
				if fieldValue == v {
					return true
				}
			}
		}
		return false
	case "not_in":
		if slice, ok := condition.Value.([]interface{}); ok {
			for _, v := range slice {
				if fieldValue == v {
					return false
				}
			}
		}
		return true
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", condition.Value))
	case "not_contains":
		return !strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", condition.Value))
	}

	return true
}

func (e *AdvancedEngine) evaluateEdgeCondition(condition types.EdgeCondition, data map[string]interface{}) bool {
	switch condition.Type {
	case "always":
		return true
	case "field_value":
		return e.evaluateCondition(ActivationCondition{
			Field:    condition.Field,
			Operator: condition.Operator,
			Value:    condition.Value,
		}, data)
	case "custom":
		// Implement custom rule evaluation
		return e.evaluateCustomRule(condition.CustomRule, data)
	default:
		return false
	}
}

func (e *AdvancedEngine) evaluateCustomRule(rule string, data map[string]interface{}) bool {
	// Implement custom rule evaluation logic
	// This could be extended to support a rule engine like OPA
	return true
}

func (e *AdvancedEngine) isEdgeDisabled(edgeID string, disabledEdges []string) bool {
	for _, disabled := range disabledEdges {
		if disabled == edgeID {
			return true
		}
	}
	return false
}

func (e *AdvancedEngine) isNodeExcluded(nodeID string, excludedNodes []string) bool {
	for _, excluded := range excludedNodes {
		if excluded == nodeID {
			return true
		}
	}
	return false
}

func (e *AdvancedEngine) isNodeVisited(nodeID string, visitedNodes []string) bool {
	for _, visited := range visitedNodes {
		if visited == nodeID {
			return true
		}
	}
	return false
}

func (e *AdvancedEngine) getCurrentTimestamp() int64 {
	// Return current timestamp in milliseconds
	return 0 // TODO: Implement proper timestamp
}

// GetStats returns statistics about the advanced engine
func (e *AdvancedEngine) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"engine_type": "advanced_engine",
		"capabilities": []string{
			"activation_rules",
			"node_rules",
			"path_limitation",
			"conditional_traversal",
			"multiple_entry_points",
		},
	}
}

// ValidateNode validates a node's data against its validation rules
func (e *AdvancedEngine) ValidateNode(ctx context.Context, node *types.Node, data map[string]interface{}) *types.ValidationResult {
	// Create a basic engine to reuse validation logic
	basicEngine := NewEngine(e.logger)
	return basicEngine.ValidateNode(ctx, node, data)
}

// CanGoBack checks if the user can go back from the current node
func (e *AdvancedEngine) CanGoBack(ctx context.Context, graph *types.Graph, currentNodeID string) bool {
	// Check if there are any edges leading to the current node
	for _, edge := range graph.Edges {
		if edge.ToNodeID == currentNodeID {
			return true
		}
	}
	return false
}
