package types

import (
	"time"

	"github.com/google/uuid"
)

// NodeType represents the type of onboarding step
type NodeType string

const (
	NodeTypeStart      NodeType = "start"
	NodeTypeInput      NodeType = "input"
	NodeTypeValidation NodeType = "validation"
	NodeTypeDecision   NodeType = "decision"
	NodeTypeEnd        NodeType = "end"
)

// Node represents a step in the onboarding process
type Node struct {
	ID          string                 `json:"id"`
	Type        NodeType               `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Fields      []Field                `json:"fields"`
	Validation  ValidationRules        `json:"validation"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Field represents an input field in a node
type Field struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       FieldType              `json:"type"`
	Required   bool                   `json:"required"`
	Options    []string               `json:"options,omitempty"`
	Validation FieldValidation        `json:"validation"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// FieldType represents the type of input field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeEmail    FieldType = "email"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeSelect   FieldType = "select"
	FieldTypeFile     FieldType = "file"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeRadio    FieldType = "radio"
)

// FieldValidation holds validation rules for a field
type FieldValidation struct {
	MinLength   int      `json:"min_length,omitempty"`
	MaxLength   int      `json:"max_length,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	MinValue    *int     `json:"min_value,omitempty"`
	MaxValue    *int     `json:"max_value,omitempty"`
	CustomRules []string `json:"custom_rules,omitempty"`
}

// ValidationRules holds validation rules for a node
type ValidationRules struct {
	RequiredFields []string              `json:"required_fields"`
	CustomRules    []string              `json:"custom_rules"`
	Conditions     []ValidationCondition `json:"conditions"`
}

// ValidationCondition represents a conditional validation rule
type ValidationCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, etc.
	Value    interface{} `json:"value"`
	Rule     string      `json:"rule"`
}

// Edge represents a connection between nodes
type Edge struct {
	ID         string                 `json:"id"`
	FromNodeID string                 `json:"from_node_id"`
	ToNodeID   string                 `json:"to_node_id"`
	Condition  EdgeCondition          `json:"condition"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// EdgeCondition represents the condition for traversing an edge
type EdgeCondition struct {
	Type       string                 `json:"type"` // always, field_value, custom
	Field      string                 `json:"field,omitempty"`
	Operator   string                 `json:"operator,omitempty"`
	Value      interface{}            `json:"value,omitempty"`
	CustomRule string                 `json:"custom_rule,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Graph represents the complete onboarding flow
type Graph struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Nodes       map[string]*Node       `json:"nodes"`
	Edges       map[string]*Edge       `json:"edges"`
	StartNodeID string                 `json:"start_node_id"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Session represents an active onboarding session
type Session struct {
	ID            string                 `json:"id"`
	UserID        string                 `json:"user_id"`
	GraphID       string                 `json:"graph_id"`
	CurrentNodeID string                 `json:"current_node_id"`
	Data          map[string]interface{} `json:"data"`
	History       []SessionStep          `json:"history"`
	Status        SessionStatus          `json:"status"`
	RetryCount    int                    `json:"retry_count"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	DynamicState  *DynamicSessionState   `json:"dynamic_state,omitempty"`
}

// DynamicSessionState represents the persistent state of dynamic nodes
type DynamicSessionState struct {
	BusinessType     string                    `json:"business_type"`
	NodeStatuses     map[string]NodeStatusInfo `json:"node_statuses"`
	LastEvaluatedAt  time.Time                 `json:"last_evaluated_at"`
	CompletionStatus map[string]interface{}    `json:"completion_status"`
}

// NodeStatusInfo represents the persistent status information for a node
type NodeStatusInfo struct {
	Status        string                 `json:"status"`
	InitialStatus string                 `json:"initial_status"`
	Dependencies  []DependencyInfo       `json:"dependencies"`
	LastUpdatedAt time.Time              `json:"last_updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DependencyInfo represents a persistent dependency
type DependencyInfo struct {
	FieldID      string      `json:"field_id"`
	Operator     string      `json:"operator"`
	Value        interface{} `json:"value"`
	Condition    string      `json:"condition"`
	BusinessType string      `json:"business_type,omitempty"`
}

// SessionStatus represents the status of an onboarding session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusPaused    SessionStatus = "paused"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusFailed    SessionStatus = "failed"
	SessionStatusExpired   SessionStatus = "expired"
)

// SessionStep represents a step taken in the onboarding process
type SessionStep struct {
	ID        string                 `json:"id"`
	NodeID    string                 `json:"node_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"` // forward, backward, retry
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Valid    bool                   `json:"valid"`
	Errors   []ValidationError      `json:"errors"`
	Warnings []ValidationWarning    `json:"warnings"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NextStepResult represents the result of determining the next step
type NextStepResult struct {
	NextNodeID     string                 `json:"next_node_id"`
	AvailablePaths []string               `json:"available_paths"`
	CanGoBack      bool                   `json:"can_go_back"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewSession creates a new onboarding session
func NewSession(userID, graphID string) *Session {
	return &Session{
		ID:         uuid.New().String(),
		UserID:     userID,
		GraphID:    graphID,
		Data:       make(map[string]interface{}),
		History:    make([]SessionStep, 0),
		Status:     SessionStatusActive,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// NewNode creates a new node
func NewNode(nodeType NodeType, name, description string) *Node {
	return &Node{
		ID:          uuid.New().String(),
		Type:        nodeType,
		Name:        name,
		Description: description,
		Fields:      make([]Field, 0),
		Validation:  ValidationRules{},
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewEdge creates a new edge
func NewEdge(fromNodeID, toNodeID string, condition EdgeCondition) *Edge {
	return &Edge{
		ID:         uuid.New().String(),
		FromNodeID: fromNodeID,
		ToNodeID:   toNodeID,
		Condition:  condition,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}
}

// NewGraph creates a new graph
func NewGraph(name, description string) *Graph {
	return &Graph{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Version:     "1.0.0",
		Nodes:       make(map[string]*Node),
		Edges:       make(map[string]*Edge),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
