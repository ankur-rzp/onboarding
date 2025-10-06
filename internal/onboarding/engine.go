package onboarding

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// RuleGroup represents a group of rules that define a complete path to onboarding completion
type RuleGroup struct {
	ID                string                          `json:"id"`
	Name              string                          `json:"name"`
	Description       string                          `json:"description"`
	RequiredNodes     []string                        `json:"required_nodes"`     // Node IDs that must be completed for this rule group
	RequiredFields    map[string][]string             `json:"required_fields"`    // Node ID -> []Field IDs that must be filled
	ConditionalFields map[string]ConditionalFieldRule `json:"conditional_fields"` // Conditional field requirements
}

// ConditionalFieldRule defines when a field is required based on other field values
type ConditionalFieldRule struct {
	NodeID      string `json:"node_id"`
	FieldID     string `json:"field_id"`
	Condition   string `json:"condition"` // Field to check
	Operator    string `json:"operator"`  // eq, ne, etc.
	Value       string `json:"value"`     // Value to compare against
	Description string `json:"description"`
}

// Engine handles graph traversal and validation
type Engine struct {
	logger *logrus.Logger
}

// NewEngine creates a new graph engine
func NewEngine(logger *logrus.Logger) *Engine {
	return &Engine{
		logger: logger,
	}
}

// ValidateNode validates a node's data against its validation rules
func (e *Engine) ValidateNode(ctx context.Context, node *Node, data map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
		Metadata: make(map[string]interface{}),
	}

	// Validate required fields
	for _, fieldName := range node.Validation.RequiredFields {
		if _, exists := data[fieldName]; !exists || data[fieldName] == nil || data[fieldName] == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field %s is required", fieldName),
				Code:    "REQUIRED_FIELD_MISSING",
			})
		}
	}

	// Validate conditionally required fields based on conditions
	for _, condition := range node.Validation.Conditions {
		// Only validate conditions that are actually met
		if e.evaluateCondition(condition, data) {
			// Condition is true, so check if the required fields for this condition are present
			requiredFields := e.getRequiredFieldsForCondition(condition, node)
			for _, fieldName := range requiredFields {
				if _, exists := data[fieldName]; !exists || data[fieldName] == nil || data[fieldName] == "" {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fieldName,
						Message: fmt.Sprintf("Field %s is required when %s", fieldName, condition.Rule),
						Code:    "CONDITIONAL_FIELD_MISSING",
					})
				}
			}
		}
		// If condition is not met, we don't validate it - this is the key fix
	}

	// Validate individual fields
	for _, field := range node.Fields {
		if value, exists := data[field.ID]; exists && value != nil {
			// Skip validation for optional fields that are empty
			valueStr := fmt.Sprintf("%v", value)
			if !field.Required && valueStr == "" {
				continue
			}

			fieldResult := e.validateField(field, value)
			if !fieldResult.Valid {
				result.Valid = false
				result.Errors = append(result.Errors, fieldResult.Errors...)
			}
			result.Warnings = append(result.Warnings, fieldResult.Warnings...)
		}
	}

	// Validate custom rules
	for _, rule := range node.Validation.CustomRules {
		if !e.validateCustomRule(rule, data) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "",
				Message: fmt.Sprintf("Custom validation rule failed: %s", rule),
				Code:    "CUSTOM_RULE_FAILED",
			})
		}
	}

	return result
}

// validateField validates a single field
func (e *Engine) validateField(field Field, value interface{}) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
	}

	valueStr := fmt.Sprintf("%v", value)

	// Length validation
	if field.Validation.MinLength > 0 && len(valueStr) < field.Validation.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   field.ID,
			Message: fmt.Sprintf("Field %s must be at least %d characters long", field.ID, field.Validation.MinLength),
			Code:    "MIN_LENGTH_VIOLATION",
		})
	}

	if field.Validation.MaxLength > 0 && len(valueStr) > field.Validation.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   field.ID,
			Message: fmt.Sprintf("Field %s must be at most %d characters long", field.ID, field.Validation.MaxLength),
			Code:    "MAX_LENGTH_VIOLATION",
		})
	}

	// Pattern validation
	if field.Validation.Pattern != "" {
		matched, err := regexp.MatchString(field.Validation.Pattern, valueStr)
		if err != nil {
			e.logger.WithError(err).Error("Invalid regex pattern")
		} else if !matched {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("Field %s does not match required pattern", field.ID),
				Code:    "PATTERN_MISMATCH",
			})
		}
	}

	// Numeric validation
	if field.Type == FieldTypeNumber {
		if num, err := strconv.Atoi(valueStr); err == nil {
			if field.Validation.MinValue != nil && num < *field.Validation.MinValue {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   field.ID,
					Message: fmt.Sprintf("Field %s must be at least %d", field.ID, *field.Validation.MinValue),
					Code:    "MIN_VALUE_VIOLATION",
				})
			}
			if field.Validation.MaxValue != nil && num > *field.Validation.MaxValue {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   field.ID,
					Message: fmt.Sprintf("Field %s must be at most %d", field.ID, *field.Validation.MaxValue),
					Code:    "MAX_VALUE_VIOLATION",
				})
			}
		}
	}

	// Email validation
	if field.Type == FieldTypeEmail {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(valueStr) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("Field %s must be a valid email address", field.ID),
				Code:    "INVALID_EMAIL",
			})
		}
	}

	// Custom field rules
	for _, rule := range field.Validation.CustomRules {
		if !e.validateCustomRule(rule, map[string]interface{}{field.ID: value}) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field.ID,
				Message: fmt.Sprintf("Custom validation rule failed for field %s: %s", field.ID, rule),
				Code:    "CUSTOM_FIELD_RULE_FAILED",
			})
		}
	}

	return result
}

// validateCustomRule validates a custom rule
func (e *Engine) validateCustomRule(rule string, data map[string]interface{}) bool {
	// This is a simplified implementation
	// In a real system, you might use a rule engine like OPA or implement a DSL

	switch rule {
	case "pan_validation", "pan_format":
		// Indian PAN format validation
		if pan, exists := data["pan_number"]; exists {
			panStr := fmt.Sprintf("%v", pan)
			panRegex := regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)
			return panRegex.MatchString(panStr)
		}
	case "aadhaar_validation", "aadhaar_format":
		// Indian Aadhaar format validation
		if aadhaar, exists := data["aadhaar_number"]; exists {
			aadhaarStr := fmt.Sprintf("%v", aadhaar)
			aadhaarRegex := regexp.MustCompile(`^[0-9]{12}$`)
			return aadhaarRegex.MatchString(aadhaarStr)
		}
	case "gst_validation", "gst_format":
		// Indian GST format validation
		if gst, exists := data["gst_number"]; exists {
			gstStr := fmt.Sprintf("%v", gst)
			gstRegex := regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)
			return gstRegex.MatchString(gstStr)
		}
	}

	return true // Default to valid for unknown rules
}

// validateCondition validates a validation condition
func (e *Engine) validateCondition(condition ValidationCondition, data map[string]interface{}) bool {
	fieldValue, exists := data[condition.Field]
	if !exists {
		// If field doesn't exist, check if it's an optional field
		// For optional fields that don't exist, we should skip the condition validation
		// Only fail if it's a required field
		return true // Skip validation for missing optional fields
	}

	switch condition.Operator {
	case "eq":
		return fieldValue == condition.Value
	case "ne":
		return fieldValue != condition.Value
	case "gt":
		return compareNumbers(fieldValue, condition.Value) > 0
	case "lt":
		return compareNumbers(fieldValue, condition.Value) < 0
	case "gte":
		return compareNumbers(fieldValue, condition.Value) >= 0
	case "lte":
		return compareNumbers(fieldValue, condition.Value) <= 0
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", condition.Value))
	case "not_contains":
		return !strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", condition.Value))
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
	}

	return true
}

// GetNextNodes returns the possible next nodes from the current node
func (e *Engine) GetNextNodes(ctx context.Context, graph *Graph, currentNodeID string, sessionData map[string]interface{}) []*Node {
	nextNodes := make([]*Node, 0)

	for _, edge := range graph.Edges {
		if edge.FromNodeID == currentNodeID {
			if e.evaluateEdgeCondition(edge.Condition, sessionData) {
				if nextNode, exists := graph.Nodes[edge.ToNodeID]; exists {
					nextNodes = append(nextNodes, nextNode)
				}
			}
		}
	}

	return nextNodes
}

// ValidatePathCompleteness checks if all required nodes have meaningful data filled
func (e *Engine) ValidatePathCompleteness(ctx context.Context, graph *Graph, currentNodeID string, sessionData map[string]interface{}, sessionHistory []SessionStep) (bool, []string) {
	// Check if this is the production onboarding graph
	businessType, hasBusinessType := sessionData["business_type"]
	if hasBusinessType {
		// For production onboarding, validate business-type specific requirements
		return e.ValidateProductionOnboardingCompleteness(ctx, graph, currentNodeID, sessionData, businessType)
	}

	// Legacy logic for unified onboarding graph
	// Get all nodes that are required for activation based on the current session data
	requiredNodes := e.getRequiredNodesForActivation(ctx, graph, sessionData)

	// Check which required nodes have ALL their required fields filled with valid data
	completedNodes := make(map[string]bool)

	// Check each required node to see if it has all required fields filled
	for _, nodeID := range requiredNodes {
		node, exists := graph.Nodes[nodeID]
		if !exists {
			continue
		}

		// Check if this node has ALL required fields filled with valid data
		allRequiredFieldsFilled := true
		missingFields := make([]string, 0)

		for _, field := range node.Fields {
			if field.Required {
				value, exists := sessionData[field.ID]
				if !exists || value == nil || value == "" {
					allRequiredFieldsFilled = false
					missingFields = append(missingFields, field.ID)
					continue
				}

				// Check if it's not just empty string or whitespace
				if strValue, ok := value.(string); ok {
					if strings.TrimSpace(strValue) == "" {
						allRequiredFieldsFilled = false
						missingFields = append(missingFields, field.ID)
						continue
					}
				}
			}
		}

		if allRequiredFieldsFilled {
			completedNodes[nodeID] = true
			e.logger.WithFields(logrus.Fields{
				"node_id":   nodeID,
				"node_name": node.Name,
			}).Debug("Node marked as completed - all required fields filled")
		} else {
			e.logger.WithFields(logrus.Fields{
				"node_id":        nodeID,
				"node_name":      node.Name,
				"missing_fields": missingFields,
			}).Debug("Node NOT completed - missing required fields")
		}
	}

	// Find missing required nodes
	missingNodes := make([]string, 0)
	for _, nodeID := range requiredNodes {
		if !completedNodes[nodeID] {
			if node, exists := graph.Nodes[nodeID]; exists {
				missingNodes = append(missingNodes, node.Name)
			}
		}
	}

	return len(missingNodes) == 0, missingNodes
}

// getRequiredNodesForActivation returns all nodes that must have data for activation based on current session data
func (e *Engine) getRequiredNodesForActivation(ctx context.Context, graph *Graph, sessionData map[string]interface{}) []string {
	requiredNodes := make([]string, 0)

	// Check if this is the production onboarding graph by looking for business_type
	_, hasBusinessType := sessionData["business_type"]
	if hasBusinessType {
		// This is the production onboarding graph - all nodes are required in sequence
		// except for conditional nodes that depend on business type
		requiredNodes = []string{
			graph.StartNodeID, // Business Type Selection (always required)
		}

		// Add all required nodes for production onboarding
		for nodeID, node := range graph.Nodes {
			if node.Type == "input" && nodeID != graph.StartNodeID {
				// All these nodes are required for production onboarding
				if node.Name == "PAN Number" ||
					node.Name == "Payment Channel" ||
					node.Name == "MCC & Policy Verification" ||
					node.Name == "Business Document" ||
					node.Name == "Authorised Signatory Details" ||
					node.Name == "Bank Account Details" ||
					node.Name == "Business Information" {
					requiredNodes = append(requiredNodes, nodeID)
				}

				// BMC Document is optional for all business types
				if node.Name == "BMC Document" {
					// Only add if required based on subcategory (for now, we'll make it optional)
					// In a real implementation, you'd check the subcategory from session data
				}
			}
		}

		return requiredNodes
	}

	// Legacy logic for unified onboarding graph
	userType, hasUserType := sessionData["user_type"]
	if !hasUserType {
		// If no user type, we can't determine requirements
		return requiredNodes
	}

	// Define required nodes based on user type and business rules
	if userType == "individual" {
		// For individuals: User Type Selection, Personal Information, Contact Information, Identity Documents, Bank Details, Document Upload
		requiredNodes = []string{
			graph.StartNodeID, // User Type Selection (always required)
		}

		// Add other required nodes for individuals - must be the CORRECT nodes for individual users
		for nodeID, node := range graph.Nodes {
			if node.Type == "input" && nodeID != graph.StartNodeID {
				// For individuals, we need Personal Information (NOT Company Information)
				if node.Name == "Personal Information" ||
					node.Name == "Contact Information" ||
					node.Name == "Identity Documents" ||
					node.Name == "Bank Details" ||
					node.Name == "Document Upload" {
					requiredNodes = append(requiredNodes, nodeID)
				}
			}
		}
	} else if userType == "company" {
		// For companies: User Type Selection, Business Type, Company Information, Contact Information, Identity Documents, Tax Information, Bank Details, Document Upload
		requiredNodes = []string{
			graph.StartNodeID, // User Type Selection (always required)
		}

		// Add other required nodes for companies - must be the CORRECT nodes for company users
		for nodeID, node := range graph.Nodes {
			if node.Type == "input" && nodeID != graph.StartNodeID {
				// For companies, we need Company Information (NOT Personal Information)
				if node.Name == "Business Type" ||
					node.Name == "Company Information" ||
					node.Name == "Contact Information" ||
					node.Name == "Identity Documents" ||
					node.Name == "Tax Information" ||
					node.Name == "Bank Details" ||
					node.Name == "Document Upload" {
					requiredNodes = append(requiredNodes, nodeID)
				}
			}
		}
	}

	return requiredNodes
}

// getRequiredNodesForPath returns all nodes that must be completed before reaching the target node
func (e *Engine) getRequiredNodesForPath(ctx context.Context, graph *Graph, targetNodeID string, sessionData map[string]interface{}) []string {
	requiredNodes := make([]string, 0)
	visited := make(map[string]bool)

	// Start from the start node and find all paths to the target node
	e.findRequiredNodesRecursive(ctx, graph, graph.StartNodeID, targetNodeID, sessionData, visited, &requiredNodes)

	return requiredNodes
}

// findRequiredNodesRecursive recursively finds all nodes that must be completed to reach the target
func (e *Engine) findRequiredNodesRecursive(ctx context.Context, graph *Graph, currentNodeID, targetNodeID string, sessionData map[string]interface{}, visited map[string]bool, requiredNodes *[]string) {
	if currentNodeID == targetNodeID {
		return // Reached target
	}

	if visited[currentNodeID] {
		return // Already visited this path
	}

	visited[currentNodeID] = true

	// Add current node to required nodes (except start node)
	if currentNodeID != graph.StartNodeID {
		*requiredNodes = append(*requiredNodes, currentNodeID)
	}

	// Check all outgoing edges from current node
	for _, edge := range graph.Edges {
		if edge.FromNodeID == currentNodeID {
			if e.evaluateEdgeCondition(edge.Condition, sessionData) {
				// This edge is valid, continue the path
				e.findRequiredNodesRecursive(ctx, graph, edge.ToNodeID, targetNodeID, sessionData, visited, requiredNodes)
			}
		}
	}
}

// evaluateEdgeCondition evaluates whether an edge condition is met
func (e *Engine) evaluateEdgeCondition(condition EdgeCondition, data map[string]interface{}) bool {
	switch condition.Type {
	case "always":
		return true
	case "field_value":
		return e.validateCondition(ValidationCondition{
			Field:    condition.Field,
			Operator: condition.Operator,
			Value:    condition.Value,
		}, data)
	case "custom":
		return e.validateCustomRule(condition.CustomRule, data)
	default:
		return false
	}
}

// CanGoBack checks if the user can go back from the current node
func (e *Engine) CanGoBack(ctx context.Context, graph *Graph, currentNodeID string) bool {
	// Check if there are any edges leading to the current node
	for _, edge := range graph.Edges {
		if edge.ToNodeID == currentNodeID {
			return true
		}
	}
	return false
}

// GetPreviousNodes returns the possible previous nodes
func (e *Engine) GetPreviousNodes(ctx context.Context, graph *Graph, currentNodeID string) []*Node {
	previousNodes := make([]*Node, 0)

	for _, edge := range graph.Edges {
		if edge.ToNodeID == currentNodeID {
			if previousNode, exists := graph.Nodes[edge.FromNodeID]; exists {
				previousNodes = append(previousNodes, previousNode)
			}
		}
	}

	return previousNodes
}

// compareNumbers compares two values as numbers
func compareNumbers(a, b interface{}) int {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	aNum, errA := strconv.ParseFloat(aStr, 64)
	bNum, errB := strconv.ParseFloat(bStr, 64)

	if errA != nil || errB != nil {
		// If either is not a number, do string comparison
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0
	}

	if aNum < bNum {
		return -1
	} else if aNum > bNum {
		return 1
	}
	return 0
}

// GetFirstMissingNode returns the ID of the first missing node that needs to be completed
func (e *Engine) GetFirstMissingNode(ctx context.Context, graph *Graph, missingNodeNames []string, sessionData map[string]interface{}) string {
	// Get user type to determine the correct order
	userType, hasUserType := sessionData["user_type"]
	if !hasUserType {
		return ""
	}

	// Define the preferred order for each user type
	var preferredOrder []string
	if userType == "individual" {
		preferredOrder = []string{
			"User Type Selection",
			"Personal Information",
			"Contact Information",
			"Identity Documents",
			"Bank Details",
			"Document Upload",
		}
	} else if userType == "company" {
		preferredOrder = []string{
			"User Type Selection",
			"Business Type",
			"Company Information",
			"Contact Information",
			"Identity Documents",
			"Tax Information",
			"Bank Details",
			"Document Upload",
		}
	}

	// Find the first missing node in the preferred order
	for _, preferredNodeName := range preferredOrder {
		for _, missingNodeName := range missingNodeNames {
			if preferredNodeName == missingNodeName {
				// Find the node ID for this name
				for nodeID, node := range graph.Nodes {
					if node.Name == missingNodeName {
						return nodeID
					}
				}
			}
		}
	}

	// Fallback: return the first missing node found
	if len(missingNodeNames) > 0 {
		for nodeID, node := range graph.Nodes {
			if node.Name == missingNodeNames[0] {
				return nodeID
			}
		}
	}

	return ""
}

// evaluateCondition evaluates a validation condition against the provided data
func (e *Engine) evaluateCondition(condition ValidationCondition, data map[string]interface{}) bool {
	fieldValue, exists := data[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "eq":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", condition.Value)
	case "ne":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", condition.Value)
	case "in":
		if values, ok := condition.Value.([]string); ok {
			fieldStr := fmt.Sprintf("%v", fieldValue)
			for _, v := range values {
				if fieldStr == v {
					return true
				}
			}
		}
		return false
	case "not_in":
		if values, ok := condition.Value.([]string); ok {
			fieldStr := fmt.Sprintf("%v", fieldValue)
			for _, v := range values {
				if fieldStr == v {
					return false
				}
			}
		}
		return true
	default:
		e.logger.WithField("operator", condition.Operator).Warn("Unknown condition operator")
		return false
	}
}

// getRequiredFieldsForCondition returns the fields that should be required when a condition is met
func (e *Engine) getRequiredFieldsForCondition(condition ValidationCondition, node *Node) []string {
	requiredFields := make([]string, 0)

	// Based on the condition, determine which fields should be required
	switch condition.Field {
	case "payment_channel":
		if condition.Operator == "eq" {
			if condition.Value == "website" {
				requiredFields = append(requiredFields, "website_url")
			} else if condition.Value == "app" {
				requiredFields = append(requiredFields, "android_url", "ios_url")
			}
		}
	case "business_type":
		// For business type conditions, we'll handle this in the business document validation
		// This is more complex and would need to check the specific business type
	}

	return requiredFields
}

// ValidateProductionOnboardingCompleteness validates that any rule group passes (all required nodes filled)
func (e *Engine) ValidateProductionOnboardingCompleteness(ctx context.Context, graph *Graph, currentNodeID string, sessionData map[string]interface{}, businessType interface{}) (bool, []string) {
	// Get business type as string
	businessTypeStr := fmt.Sprintf("%v", businessType)

	// Get all rule groups for this business type
	ruleGroups := e.GetBusinessTypeRuleGroups(businessTypeStr)

	// Check if ANY rule group passes (all its required nodes are filled)
	for _, ruleGroup := range ruleGroups {
		ruleGroupPassed, _ := e.EvaluateRuleGroup(ruleGroup, sessionData)
		if ruleGroupPassed {
			// This rule group passes - user is eligible for completion
			e.logger.WithFields(logrus.Fields{
				"business_type": businessTypeStr,
				"rule_group":    ruleGroup.Name,
				"current_node":  currentNodeID,
			}).Info("Rule group passed - eligible for completion")
			return true, []string{}
		}
	}

	// No rule group passed - collect all missing requirements for logging
	allMissingRequirements := make([]string, 0)
	for _, ruleGroup := range ruleGroups {
		_, missingNodes := e.EvaluateRuleGroup(ruleGroup, sessionData)
		allMissingRequirements = append(allMissingRequirements, fmt.Sprintf("Rule Group '%s': %v", ruleGroup.Name, missingNodes))
	}

	// Check if we're at the completion node
	currentNode, exists := graph.Nodes[currentNodeID]
	if !exists {
		return false, []string{"Current node not found"}
	}

	// If we're at the completion node but no rule group passes, that's an error
	if currentNode.Type == "end" {
		return false, allMissingRequirements
	}

	// If we're not at the completion node, we're not complete yet
	return false, allMissingRequirements
}

// GetBusinessTypeRuleGroups returns all rule groups for a business type
func (e *Engine) GetBusinessTypeRuleGroups(businessType string) []RuleGroup {
	allRuleGroups := e.getAllRuleGroups()
	if ruleGroups, exists := allRuleGroups[businessType]; exists {
		return ruleGroups
	}

	// Return empty rule groups if not found
	return []RuleGroup{}
}

// getAllRuleGroups defines rule groups for each business type - multiple paths to completion
func (e *Engine) getAllRuleGroups() map[string][]RuleGroup {
	ruleGroups := make(map[string][]RuleGroup)

	// Individual business type rule groups
	ruleGroups["individual"] = []RuleGroup{
		{
			ID:          "individual_basic_path",
			Name:        "Individual Basic Path",
			Description: "Basic individual onboarding with essential fields only",
			RequiredNodes: []string{
				"business_type_selection",
				"payment_channel",
				"business_details",
			},
			RequiredFields: map[string][]string{
				"business_type_selection": {"business_type"},
				"payment_channel":         {"payment_channel"},
				"business_details":        {"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
			},
			ConditionalFields: map[string]ConditionalFieldRule{
				"website_url_required": {
					NodeID:      "payment_channel",
					FieldID:     "website_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "website",
					Description: "Website URL is required when payment channel is website",
				},
				"app_urls_required": {
					NodeID:      "payment_channel",
					FieldID:     "android_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "app",
					Description: "Android and iOS URLs are required when payment channel is app",
				},
			},
		},
		{
			ID:          "individual_complete_path",
			Name:        "Individual Complete Path",
			Description: "Complete individual onboarding with all fields including documents",
			RequiredNodes: []string{
				"business_type_selection",
				"payment_channel",
				"mcc_policy_verification",
				"business_details",
			},
			RequiredFields: map[string][]string{
				"business_type_selection": {"business_type"},
				"payment_channel":         {"payment_channel"},
				"mcc_policy_verification": {"subcategory"},
				"business_details":        {"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
			},
			ConditionalFields: map[string]ConditionalFieldRule{
				"website_url_required": {
					NodeID:      "payment_channel",
					FieldID:     "website_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "website",
					Description: "Website URL is required when payment channel is website",
				},
				"app_urls_required": {
					NodeID:      "payment_channel",
					FieldID:     "android_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "app",
					Description: "Android and iOS URLs are required when payment channel is app",
				},
			},
		},
	}

	// Proprietorship business type rule groups
	ruleGroups["proprietorship"] = []RuleGroup{
		{
			ID:          "proprietorship_basic_path",
			Name:        "Proprietorship Basic Path",
			Description: "Basic proprietorship onboarding with essential fields only",
			RequiredNodes: []string{
				"business_type_selection",
				"payment_channel",
				"business_details",
			},
			RequiredFields: map[string][]string{
				"business_type_selection": {"business_type"},
				"payment_channel":         {"payment_channel"},
				"business_details":        {"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
			},
			ConditionalFields: map[string]ConditionalFieldRule{
				"website_url_required": {
					NodeID:      "payment_channel",
					FieldID:     "website_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "website",
					Description: "Website URL is required when payment channel is website",
				},
				"app_urls_required": {
					NodeID:      "payment_channel",
					FieldID:     "android_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "app",
					Description: "Android and iOS URLs are required when payment channel is app",
				},
			},
		},
		{
			ID:          "proprietorship_complete_path",
			Name:        "Proprietorship Complete Path",
			Description: "Complete proprietorship onboarding with all fields including documents",
			RequiredNodes: []string{
				"business_type_selection",
				"payment_channel",
				"mcc_policy_verification",
				"business_details",
			},
			RequiredFields: map[string][]string{
				"business_type_selection": {"business_type"},
				"payment_channel":         {"payment_channel"},
				"mcc_policy_verification": {"subcategory"},
				"business_details":        {"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
			},
			ConditionalFields: map[string]ConditionalFieldRule{
				"website_url_required": {
					NodeID:      "payment_channel",
					FieldID:     "website_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "website",
					Description: "Website URL is required when payment channel is website",
				},
				"app_urls_required": {
					NodeID:      "payment_channel",
					FieldID:     "android_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "app",
					Description: "Android and iOS URLs are required when payment channel is app",
				},
			},
		},
	}

	// Private Limited business type rule groups
	ruleGroups["private_limited"] = []RuleGroup{
		{
			ID:          "private_limited_basic_path",
			Name:        "Private Limited Basic Path",
			Description: "Basic private limited onboarding with essential fields only",
			RequiredNodes: []string{
				"business_type_selection",
				"payment_channel",
				"business_details",
			},
			RequiredFields: map[string][]string{
				"business_type_selection": {"business_type"},
				"payment_channel":         {"payment_channel"},
				"business_details":        {"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
			},
			ConditionalFields: map[string]ConditionalFieldRule{
				"website_url_required": {
					NodeID:      "payment_channel",
					FieldID:     "website_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "website",
					Description: "Website URL is required when payment channel is website",
				},
				"app_urls_required": {
					NodeID:      "payment_channel",
					FieldID:     "android_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "app",
					Description: "Android and iOS URLs are required when payment channel is app",
				},
			},
		},
		{
			ID:          "private_limited_complete_path",
			Name:        "Private Limited Complete Path",
			Description: "Complete private limited onboarding with all fields including documents",
			RequiredNodes: []string{
				"business_type_selection",
				"payment_channel",
				"mcc_policy_verification",
				"business_details",
			},
			RequiredFields: map[string][]string{
				"business_type_selection": {"business_type"},
				"payment_channel":         {"payment_channel"},
				"mcc_policy_verification": {"subcategory"},
				"business_details":        {"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
			},
			ConditionalFields: map[string]ConditionalFieldRule{
				"website_url_required": {
					NodeID:      "payment_channel",
					FieldID:     "website_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "website",
					Description: "Website URL is required when payment channel is website",
				},
				"app_urls_required": {
					NodeID:      "payment_channel",
					FieldID:     "android_url",
					Condition:   "payment_channel",
					Operator:    "eq",
					Value:       "app",
					Description: "Android and iOS URLs are required when payment channel is app",
				},
			},
		},
	}

	// Add similar rule groups for other business types...
	// For now, using the same structure for all types
	for _, businessType := range []string{"public_limited", "partnership", "llp", "trust", "society", "huf"} {
		baseRuleGroups := ruleGroups["private_limited"] // Copy the same structure
		for i := range baseRuleGroups {
			baseRuleGroups[i].ID = businessType + "_" + baseRuleGroups[i].ID
			// Convert business type to title case
			titleCase := strings.ReplaceAll(businessType, "_", " ")
			titleCase = strings.Title(titleCase)
			baseRuleGroups[i].Name = strings.Replace(baseRuleGroups[i].Name, "Private Limited", titleCase, 1)
			baseRuleGroups[i].Description = strings.Replace(baseRuleGroups[i].Description, "private limited", businessType, 1)
		}
		ruleGroups[businessType] = baseRuleGroups
	}

	return ruleGroups
}

// EvaluateRuleGroup checks if all required nodes and fields are completed according to the rule group
func (e *Engine) EvaluateRuleGroup(ruleGroup RuleGroup, sessionData map[string]interface{}) (bool, []string) {
	missingRequirements := make([]string, 0)

	// Check if all required nodes have been visited (have data)
	for _, nodeID := range ruleGroup.RequiredNodes {
		// Check if this node has been completed by looking for its key fields in session data
		requiredFields, hasRequiredFields := ruleGroup.RequiredFields[nodeID]
		if !hasRequiredFields {
			continue
		}

		for _, fieldID := range requiredFields {
			if value, exists := sessionData[fieldID]; !exists || value == nil || value == "" {
				missingRequirements = append(missingRequirements, fmt.Sprintf("Node %s: field %s is required", nodeID, fieldID))
			}
		}

		// Check conditional fields for this node
		for _, conditionalRule := range ruleGroup.ConditionalFields {
			if conditionalRule.NodeID == nodeID {
				if e.evaluateConditionalField(conditionalRule, sessionData) {
					// Condition is met, so this field is required
					if value, exists := sessionData[conditionalRule.FieldID]; !exists || value == nil || value == "" {
						missingRequirements = append(missingRequirements, fmt.Sprintf("Node %s: field %s is required (%s)", nodeID, conditionalRule.FieldID, conditionalRule.Description))
					}
				}
			}
		}
	}

	// Rule group passes if no missing requirements
	return len(missingRequirements) == 0, missingRequirements
}

// evaluateConditionalField checks if a conditional field requirement is met
func (e *Engine) evaluateConditionalField(rule ConditionalFieldRule, sessionData map[string]interface{}) bool {
	conditionValue, exists := sessionData[rule.Condition]
	if !exists {
		return false
	}

	conditionStr := fmt.Sprintf("%v", conditionValue)

	switch rule.Operator {
	case "eq":
		return conditionStr == rule.Value
	case "ne":
		return conditionStr != rule.Value
	default:
		return false
	}
}

// getBusinessTypeRequirements returns the specific requirements for each business type based on CSV rules
func (e *Engine) getBusinessTypeRequirements(businessType string) map[string]string {
	// This should match the CSV rules from the production onboarding
	requirements := map[string]string{
		// Common requirements for all business types
		"business_type":           "required",
		"pan_number":              "required",
		"pan_document":            "required",
		"payment_channel":         "required",
		"subcategory":             "required",
		"policy_pages":            "required",
		"signatory_name":          "required",
		"signatory_aadhaar_front": "required",
		"signatory_aadhaar_back":  "required",
		"bank_account_number":     "required",
		"ifsc_code":               "required",
		"bank_name":               "required",
		"business_name":           "required",
		"brand_name":              "required",
		"business_address_line1":  "required",
		"business_city":           "required",
		"business_state":          "required",
		"business_pincode":        "required",
	}

	// Business-type specific requirements
	switch businessType {
	case "proprietorship":
		requirements["msme_document"] = "required"
	case "private_limited", "public_limited":
		requirements["cin_document"] = "required"
		requirements["certificate_of_incorporation"] = "required"
	case "partnership":
		requirements["partnership_deed"] = "required"
	case "llp":
		requirements["certificate_of_incorporation"] = "required"
	case "trust":
		requirements["trust_deed"] = "required"
	case "society":
		requirements["society_registration_certificate"] = "required"
	case "huf":
		requirements["huf_deed"] = "required"
	}

	return requirements
}

// validateConditionalRequirements validates conditional requirements based on field values
func (e *Engine) validateConditionalRequirements(sessionData map[string]interface{}, businessType string) []string {
	missingRequirements := make([]string, 0)

	// Check payment channel conditional requirements
	if paymentChannel, exists := sessionData["payment_channel"]; exists {
		paymentChannelStr := fmt.Sprintf("%v", paymentChannel)
		switch paymentChannelStr {
		case "website":
			if websiteURL, exists := sessionData["website_url"]; !exists || websiteURL == nil || websiteURL == "" {
				missingRequirements = append(missingRequirements, "website_url is required when payment_channel is website")
			}
		case "app":
			if androidURL, exists := sessionData["android_url"]; !exists || androidURL == nil || androidURL == "" {
				missingRequirements = append(missingRequirements, "android_url is required when payment_channel is app")
			}
			if iosURL, exists := sessionData["ios_url"]; !exists || iosURL == nil || iosURL == "" {
				missingRequirements = append(missingRequirements, "ios_url is required when payment_channel is app")
			}
		}
	}

	// Check subcategory conditional requirements for BMC document
	if subcategory, exists := sessionData["subcategory"]; exists {
		// Add logic here for specific subcategories that require BMC document
		// For now, we'll make it optional as per CSV rules
		_ = subcategory // Use the variable to avoid warning
	}

	return missingRequirements
}
