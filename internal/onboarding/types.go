package onboarding

import (
	"onboarding-system/internal/types"
)

// Re-export types from the main types package for convenience
type Node = types.Node
type Field = types.Field
type Graph = types.Graph
type Edge = types.Edge
type Session = types.Session
type SessionStep = types.SessionStep
type SessionStatus = types.SessionStatus
type ValidationResult = types.ValidationResult
type ValidationError = types.ValidationError
type ValidationWarning = types.ValidationWarning
type ValidationCondition = types.ValidationCondition
type EdgeCondition = types.EdgeCondition
type FieldValidation = types.FieldValidation
type ValidationRules = types.ValidationRules
type NodeType = types.NodeType
type FieldType = types.FieldType
type NextStepResult = types.NextStepResult

// Re-export functions
var NewSession = types.NewSession

// Constants
const (
	NodeTypeStart      = types.NodeTypeStart
	NodeTypeInput      = types.NodeTypeInput
	NodeTypeValidation = types.NodeTypeValidation
	NodeTypeDecision   = types.NodeTypeDecision
	NodeTypeEnd        = types.NodeTypeEnd
)

const (
	FieldTypeText     = types.FieldTypeText
	FieldTypeEmail    = types.FieldTypeEmail
	FieldTypeNumber   = types.FieldTypeNumber
	FieldTypeSelect   = types.FieldTypeSelect
	FieldTypeCheckbox = types.FieldTypeCheckbox
	FieldTypeDate     = types.FieldTypeDate
	FieldTypeFile     = types.FieldTypeFile
)

const (
	SessionStatusActive    = types.SessionStatusActive
	SessionStatusPaused    = types.SessionStatusPaused
	SessionStatusCompleted = types.SessionStatusCompleted
	SessionStatusFailed    = types.SessionStatusFailed
	SessionStatusExpired   = types.SessionStatusExpired
)
