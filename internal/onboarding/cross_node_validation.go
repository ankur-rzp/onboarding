package onboarding

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"onboarding-system/internal/types"

	"github.com/sirupsen/logrus"
)

// CrossNodeValidationResult represents the result of cross-node validation
type CrossNodeValidationResult struct {
	RuleID   string                   `json:"rule_id"`
	RuleName string                   `json:"rule_name"`
	Passed   bool                     `json:"passed"`
	ErrorMsg string                   `json:"error_msg,omitempty"`
	Severity types.ValidationSeverity `json:"severity"`
	Fields   []string                 `json:"fields"`  // Fields that were validated
	Details  map[string]interface{}   `json:"details"` // Additional validation details
}

// CrossNodeValidationEngine handles cross-node validation logic
type CrossNodeValidationEngine struct {
	logger *logrus.Logger
}

// NewCrossNodeValidationEngine creates a new cross-node validation engine
func NewCrossNodeValidationEngine(logger *logrus.Logger) *CrossNodeValidationEngine {
	return &CrossNodeValidationEngine{
		logger: logger,
	}
}

// ValidateCrossNodeRules validates all cross-node rules for a given session
func (cve *CrossNodeValidationEngine) ValidateCrossNodeRules(ctx context.Context, graph *types.Graph, sessionData map[string]interface{}, businessType string) ([]CrossNodeValidationResult, error) {
	var results []CrossNodeValidationResult

	if len(graph.CrossNodeValidation) == 0 {
		cve.logger.Debug("No cross-node validation rules found")
		return results, nil
	}

	cve.logger.WithFields(logrus.Fields{
		"graph_id":      graph.ID,
		"business_type": businessType,
		"rule_count":    len(graph.CrossNodeValidation),
	}).Info("Starting cross-node validation")

	for _, rule := range graph.CrossNodeValidation {
		if !rule.Enabled {
			cve.logger.WithField("rule_id", rule.ID).Debug("Skipping disabled rule")
			continue
		}

		// Check if rule applies to current business type
		if rule.BusinessType != "" && rule.BusinessType != businessType {
			cve.logger.WithFields(logrus.Fields{
				"rule_id":               rule.ID,
				"rule_business_type":    rule.BusinessType,
				"current_business_type": businessType,
			}).Debug("Skipping rule for different business type")
			continue
		}

		result := cve.validateSingleRule(ctx, rule, sessionData, graph)
		results = append(results, result)

		cve.logger.WithFields(logrus.Fields{
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
			"passed":    result.Passed,
			"severity":  result.Severity,
		}).Debug("Cross-node validation rule completed")
	}

	cve.logger.WithFields(logrus.Fields{
		"total_rules": len(results),
		"passed":      countPassed(results),
		"failed":      countFailed(results),
	}).Info("Cross-node validation completed")

	return results, nil
}

// validateSingleRule validates a single cross-node rule
func (cve *CrossNodeValidationEngine) validateSingleRule(ctx context.Context, rule types.CrossNodeValidationRule, sessionData map[string]interface{}, graph *types.Graph) CrossNodeValidationResult {
	result := CrossNodeValidationResult{
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Severity: rule.Severity,
		Fields:   make([]string, 0),
		Details:  make(map[string]interface{}),
	}

	// Extract field values from session data
	fieldValues := make(map[string]interface{})
	fieldNames := make([]string, 0)

	for _, fieldRef := range rule.Fields {
		if value, exists := sessionData[fieldRef.FieldID]; exists {
			fieldValues[fieldRef.Alias] = value
			fieldNames = append(fieldNames, fieldRef.FieldID)

			// Also store the full field reference for debugging
			result.Details[fieldRef.Alias] = map[string]interface{}{
				"node_id":  fieldRef.NodeID,
				"field_id": fieldRef.FieldID,
				"value":    value,
			}
		} else {
			// Field not found in session data
			cve.logger.WithFields(logrus.Fields{
				"rule_id":  rule.ID,
				"node_id":  fieldRef.NodeID,
				"field_id": fieldRef.FieldID,
				"alias":    fieldRef.Alias,
			}).Warn("Field not found in session data for cross-node validation")

			// For missing fields, we might want to skip validation or treat as failed
			// This depends on the business logic - for now, we'll skip
			result.Passed = true
			result.ErrorMsg = fmt.Sprintf("Field %s not found in session data", fieldRef.FieldID)
			return result
		}
	}

	result.Fields = fieldNames

	// Validate based on condition type
	switch rule.Condition.Type {
	case "field_match":
		result.Passed = cve.validateFieldMatch(rule.Condition, fieldValues)
	case "field_contains":
		result.Passed = cve.validateFieldContains(rule.Condition, fieldValues)
	case "custom_logic":
		result.Passed = cve.validateCustomLogic(rule.Condition, fieldValues)
	default:
		cve.logger.WithFields(logrus.Fields{
			"rule_id": rule.ID,
			"type":    rule.Condition.Type,
		}).Error("Unknown cross-node validation condition type")
		result.Passed = false
		result.ErrorMsg = fmt.Sprintf("Unknown validation type: %s", rule.Condition.Type)
	}

	if !result.Passed {
		result.ErrorMsg = rule.ErrorMsg
	}

	return result
}

// validateFieldMatch validates that fields match according to the condition
func (cve *CrossNodeValidationEngine) validateFieldMatch(condition types.CrossNodeCondition, fieldValues map[string]interface{}) bool {
	if len(condition.Fields) < 2 {
		cve.logger.Error("Field match validation requires at least 2 fields")
		return false
	}

	// Get the first field value as the reference
	referenceValue, exists := fieldValues[condition.Fields[0]]
	if !exists {
		cve.logger.WithField("field", condition.Fields[0]).Error("Reference field not found")
		return false
	}

	// Compare all other fields with the reference
	for i := 1; i < len(condition.Fields); i++ {
		fieldValue, exists := fieldValues[condition.Fields[i]]
		if !exists {
			cve.logger.WithField("field", condition.Fields[i]).Error("Comparison field not found")
			return false
		}

		if !cve.compareValues(referenceValue, fieldValue, condition.Operator) {
			cve.logger.WithFields(logrus.Fields{
				"reference_field":  condition.Fields[0],
				"reference_value":  referenceValue,
				"comparison_field": condition.Fields[i],
				"comparison_value": fieldValue,
				"operator":         condition.Operator,
			}).Debug("Field match validation failed")
			return false
		}
	}

	return true
}

// validateFieldContains validates that fields contain certain values or patterns
func (cve *CrossNodeValidationEngine) validateFieldContains(condition types.CrossNodeCondition, fieldValues map[string]interface{}) bool {
	if condition.Value == nil {
		cve.logger.Error("Field contains validation requires a value to check against")
		return false
	}

	for _, fieldAlias := range condition.Fields {
		fieldValue, exists := fieldValues[fieldAlias]
		if !exists {
			cve.logger.WithField("field", fieldAlias).Error("Field not found for contains validation")
			return false
		}

		if !cve.containsValue(fieldValue, condition.Value, condition.Operator) {
			cve.logger.WithFields(logrus.Fields{
				"field":    fieldAlias,
				"value":    fieldValue,
				"contains": condition.Value,
				"operator": condition.Operator,
			}).Debug("Field contains validation failed")
			return false
		}
	}

	return true
}

// validateCustomLogic validates using custom logic (for complex cases)
func (cve *CrossNodeValidationEngine) validateCustomLogic(condition types.CrossNodeCondition, fieldValues map[string]interface{}) bool {
	// For now, implement some common custom logic patterns
	// In a real implementation, this could be extended with a scripting engine

	switch condition.Logic {
	case "business_name_matches_bank_name":
		return cve.validateBusinessNameMatchesBankName(fieldValues)
	case "pan_matches_signatory_pan":
		return cve.validatePanMatchesSignatoryPan(fieldValues)
	case "address_consistency":
		return cve.validateAddressConsistency(fieldValues)
	default:
		cve.logger.WithField("logic", condition.Logic).Error("Unknown custom logic")
		return false
	}
}

// validateBusinessNameMatchesBankName validates that business name matches bank name
func (cve *CrossNodeValidationEngine) validateBusinessNameMatchesBankName(fieldValues map[string]interface{}) bool {
	businessName, hasBusiness := fieldValues["business_name"]
	bankName, hasBank := fieldValues["bank_name"]

	if !hasBusiness || !hasBank {
		return false
	}

	businessStr := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", businessName)))
	bankStr := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", bankName)))

	// Allow partial matches (bank name can be shorter)
	return strings.Contains(bankStr, businessStr) || strings.Contains(businessStr, bankStr)
}

// validatePanMatchesSignatoryPan validates that PAN matches signatory PAN
func (cve *CrossNodeValidationEngine) validatePanMatchesSignatoryPan(fieldValues map[string]interface{}) bool {
	pan, hasPan := fieldValues["pan_number"]
	signatoryPan, hasSignatoryPan := fieldValues["signatory_pan"]

	if !hasPan || !hasSignatoryPan {
		return false
	}

	return fmt.Sprintf("%v", pan) == fmt.Sprintf("%v", signatoryPan)
}

// validateAddressConsistency validates address consistency across nodes
func (cve *CrossNodeValidationEngine) validateAddressConsistency(fieldValues map[string]interface{}) bool {
	businessCity, hasBusinessCity := fieldValues["business_city"]
	businessState, hasBusinessState := fieldValues["business_state"]
	businessPincode, hasBusinessPincode := fieldValues["business_pincode"]

	// All address fields should be present and non-empty
	if !hasBusinessCity || !hasBusinessState || !hasBusinessPincode {
		return false
	}

	cityStr := strings.TrimSpace(fmt.Sprintf("%v", businessCity))
	stateStr := strings.TrimSpace(fmt.Sprintf("%v", businessState))
	pincodeStr := strings.TrimSpace(fmt.Sprintf("%v", businessPincode))

	return cityStr != "" && stateStr != "" && pincodeStr != ""
}

// compareValues compares two values using the specified operator
func (cve *CrossNodeValidationEngine) compareValues(value1, value2 interface{}, operator string) bool {
	str1 := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", value1)))
	str2 := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", value2)))

	switch operator {
	case "eq", "equals":
		return str1 == str2
	case "ne", "not_equals":
		return str1 != str2
	case "contains":
		return strings.Contains(str1, str2) || strings.Contains(str2, str1)
	case "matches":
		// Use regex matching
		matched, err := regexp.MatchString(str2, str1)
		return err == nil && matched
	default:
		cve.logger.WithField("operator", operator).Error("Unknown comparison operator")
		return false
	}
}

// containsValue checks if a field value contains the specified value
func (cve *CrossNodeValidationEngine) containsValue(fieldValue, containsValue interface{}, operator string) bool {
	fieldStr := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", fieldValue)))
	containsStr := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", containsValue)))

	switch operator {
	case "contains":
		return strings.Contains(fieldStr, containsStr)
	case "starts_with":
		return strings.HasPrefix(fieldStr, containsStr)
	case "ends_with":
		return strings.HasSuffix(fieldStr, containsStr)
	case "matches":
		matched, err := regexp.MatchString(containsStr, fieldStr)
		return err == nil && matched
	default:
		cve.logger.WithField("operator", operator).Error("Unknown contains operator")
		return false
	}
}

// Helper functions for result analysis
func countPassed(results []CrossNodeValidationResult) int {
	count := 0
	for _, result := range results {
		if result.Passed {
			count++
		}
	}
	return count
}

func countFailed(results []CrossNodeValidationResult) int {
	count := 0
	for _, result := range results {
		if !result.Passed {
			count++
		}
	}
	return count
}

// GetValidationErrors returns only the failed validation results
func GetValidationErrors(results []CrossNodeValidationResult) []CrossNodeValidationResult {
	var errors []CrossNodeValidationResult
	for _, result := range results {
		if !result.Passed {
			errors = append(errors, result)
		}
	}
	return errors
}

// GetValidationWarnings returns only the warning validation results
func GetValidationWarnings(results []CrossNodeValidationResult) []CrossNodeValidationResult {
	var warnings []CrossNodeValidationResult
	for _, result := range results {
		if result.Severity == types.ValidationSeverityWarning {
			warnings = append(warnings, result)
		}
	}
	return warnings
}
