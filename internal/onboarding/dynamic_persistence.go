package onboarding

import (
	"fmt"
	"time"

	"onboarding-system/internal/types"

	"github.com/sirupsen/logrus"
)

// DynamicPersistenceManager handles persistence of dynamic node states
type DynamicPersistenceManager struct {
	logger *logrus.Logger
}

// NewDynamicPersistenceManager creates a new persistence manager
func NewDynamicPersistenceManager(logger *logrus.Logger) *DynamicPersistenceManager {
	return &DynamicPersistenceManager{
		logger: logger,
	}
}

// SaveDynamicState saves the current dynamic state to the session
func (dpm *DynamicPersistenceManager) SaveDynamicState(session *types.Session, dynamicGraph *DynamicGraph, businessType string) {
	if session.DynamicState == nil {
		session.DynamicState = &types.DynamicSessionState{
			BusinessType:     businessType,
			NodeStatuses:     make(map[string]types.NodeStatusInfo),
			LastEvaluatedAt:  time.Now(),
			CompletionStatus: make(map[string]interface{}),
		}
	}

	// Update business type if changed
	if session.DynamicState.BusinessType != businessType {
		session.DynamicState.BusinessType = businessType
		dpm.logger.WithFields(logrus.Fields{
			"session_id":        session.ID,
			"old_business_type": session.DynamicState.BusinessType,
			"new_business_type": businessType,
		}).Info("Business type changed, updating dynamic state")
	}

	// Save node statuses
	for nodeID, dynamicNode := range dynamicGraph.DynamicNodes {
		nodeStatusInfo := types.NodeStatusInfo{
			Status:        string(dynamicNode.Status),
			InitialStatus: string(dynamicNode.InitialStatus),
			Dependencies:  dpm.convertDependencies(dynamicNode.Dependencies),
			LastUpdatedAt: time.Now(),
			Metadata: map[string]interface{}{
				"node_name": dynamicNode.Name,
				"node_type": dynamicNode.Type,
			},
		}
		session.DynamicState.NodeStatuses[nodeID] = nodeStatusInfo
	}

	// Save completion status
	session.DynamicState.CompletionStatus = dynamicGraph.GetCompletionStatus()
	session.DynamicState.LastEvaluatedAt = time.Now()

	dpm.logger.WithFields(logrus.Fields{
		"session_id":        session.ID,
		"business_type":     businessType,
		"nodes_saved":       len(session.DynamicState.NodeStatuses),
		"completion_status": session.DynamicState.CompletionStatus,
	}).Info("Dynamic state saved to session")
}

// RestoreDynamicState restores the dynamic state from the session
func (dpm *DynamicPersistenceManager) RestoreDynamicState(session *types.Session, dynamicGraph *DynamicGraph) error {
	if session.DynamicState == nil {
		dpm.logger.WithFields(logrus.Fields{
			"session_id": session.ID,
		}).Info("No dynamic state found in session, using initial state")
		return nil
	}

	dpm.logger.WithFields(logrus.Fields{
		"session_id":       session.ID,
		"business_type":    session.DynamicState.BusinessType,
		"nodes_to_restore": len(session.DynamicState.NodeStatuses),
		"last_evaluated":   session.DynamicState.LastEvaluatedAt,
	}).Info("Restoring dynamic state from session")

	// Restore node statuses
	for nodeID, nodeStatusInfo := range session.DynamicState.NodeStatuses {
		if dynamicNode, exists := dynamicGraph.DynamicNodes[nodeID]; exists {
			// Restore status
			dynamicNode.Status = NodeStatus(nodeStatusInfo.Status)
			dynamicNode.InitialStatus = NodeStatus(nodeStatusInfo.InitialStatus)

			// Restore dependencies
			dynamicNode.Dependencies = dpm.convertToNodeDependencies(nodeStatusInfo.Dependencies)

			dpm.logger.WithFields(logrus.Fields{
				"session_id": session.ID,
				"node_id":    nodeID,
				"node_name":  dynamicNode.Name,
				"status":     dynamicNode.Status,
			}).Debug("Restored node status")
		} else {
			dpm.logger.WithFields(logrus.Fields{
				"session_id": session.ID,
				"node_id":    nodeID,
			}).Warn("Node not found in dynamic graph during restoration")
		}
	}

	// Re-evaluate dependencies based on current session data
	dpm.reEvaluateDependencies(dynamicGraph, session.Data)

	dpm.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
	}).Info("Dynamic state restoration completed")

	return nil
}

// convertDependencies converts NodeDependency to DependencyInfo
func (dpm *DynamicPersistenceManager) convertDependencies(dependencies []NodeDependency) []types.DependencyInfo {
	result := make([]types.DependencyInfo, len(dependencies))
	for i, dep := range dependencies {
		result[i] = types.DependencyInfo{
			FieldID:      dep.FieldID,
			Operator:     dep.Operator,
			Value:        dep.Value,
			Condition:    dep.Condition,
			BusinessType: dep.BusinessType,
		}
	}
	return result
}

// convertToNodeDependencies converts DependencyInfo to NodeDependency
func (dpm *DynamicPersistenceManager) convertToNodeDependencies(dependencies []types.DependencyInfo) []NodeDependency {
	result := make([]NodeDependency, len(dependencies))
	for i, dep := range dependencies {
		result[i] = NodeDependency{
			FieldID:      dep.FieldID,
			Operator:     dep.Operator,
			Value:        dep.Value,
			Condition:    dep.Condition,
			BusinessType: dep.BusinessType,
		}
	}
	return result
}

// reEvaluateDependencies re-evaluates dependencies based on current session data
func (dpm *DynamicPersistenceManager) reEvaluateDependencies(dynamicGraph *DynamicGraph, sessionData map[string]interface{}) {
	engine := &DynamicEngine{}

	for nodeID, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Status == NodeStatusDependent {
			newStatus := engine.EvaluateNodeDependencies(dynamicNode, sessionData)
			if newStatus != dynamicNode.Status {
				dpm.logger.WithFields(logrus.Fields{
					"node_id":    nodeID,
					"node_name":  dynamicNode.Name,
					"old_status": dynamicNode.Status,
					"new_status": newStatus,
				}).Info("Node status changed during re-evaluation")

				dynamicNode.Status = newStatus
			}
		}
	}
}

// MigrateSessionToDynamic migrates an existing session to use dynamic state
func (dpm *DynamicPersistenceManager) MigrateSessionToDynamic(session *types.Session, dynamicGraph *DynamicGraph, businessType string) error {
	if session.DynamicState != nil {
		// Already migrated
		return nil
	}

	dpm.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
	}).Info("Migrating session to dynamic state")

	// Initialize dynamic state
	session.DynamicState = &types.DynamicSessionState{
		BusinessType:     businessType,
		NodeStatuses:     make(map[string]types.NodeStatusInfo),
		LastEvaluatedAt:  time.Now(),
		CompletionStatus: make(map[string]interface{}),
	}

	// Determine business type from session data if not provided
	if businessType == "" {
		if bt, exists := session.Data["business_type"]; exists {
			businessType = fmt.Sprintf("%v", bt)
		} else {
			businessType = "individual" // default
		}
	}

	// Save current state
	dpm.SaveDynamicState(session, dynamicGraph, businessType)

	dpm.logger.WithFields(logrus.Fields{
		"session_id":    session.ID,
		"business_type": businessType,
	}).Info("Session migration to dynamic state completed")

	return nil
}

// GetDynamicStateSummary returns a summary of the dynamic state
func (dpm *DynamicPersistenceManager) GetDynamicStateSummary(session *types.Session) map[string]interface{} {
	if session.DynamicState == nil {
		return map[string]interface{}{
			"has_dynamic_state": false,
		}
	}

	summary := map[string]interface{}{
		"has_dynamic_state": true,
		"business_type":     session.DynamicState.BusinessType,
		"last_evaluated_at": session.DynamicState.LastEvaluatedAt,
		"total_nodes":       len(session.DynamicState.NodeStatuses),
		"completion_status": session.DynamicState.CompletionStatus,
	}

	// Count nodes by status
	statusCounts := make(map[string]int)
	for _, nodeStatus := range session.DynamicState.NodeStatuses {
		statusCounts[nodeStatus.Status]++
	}
	summary["status_counts"] = statusCounts

	return summary
}

// ValidateDynamicState validates the integrity of the dynamic state
func (dpm *DynamicPersistenceManager) ValidateDynamicState(session *types.Session, dynamicGraph *DynamicGraph) []string {
	var issues []string

	if session.DynamicState == nil {
		issues = append(issues, "No dynamic state found")
		return issues
	}

	// Check if all nodes in dynamic graph have corresponding state
	for nodeID := range dynamicGraph.DynamicNodes {
		if _, exists := session.DynamicState.NodeStatuses[nodeID]; !exists {
			issues = append(issues, fmt.Sprintf("Missing state for node: %s", nodeID))
		}
	}

	// Check if there are orphaned states
	for nodeID := range session.DynamicState.NodeStatuses {
		if _, exists := dynamicGraph.DynamicNodes[nodeID]; !exists {
			issues = append(issues, fmt.Sprintf("Orphaned state for node: %s", nodeID))
		}
	}

	// Validate business type consistency
	if session.DynamicState.BusinessType != "" {
		if sessionDataBT, exists := session.Data["business_type"]; exists {
			if fmt.Sprintf("%v", sessionDataBT) != session.DynamicState.BusinessType {
				issues = append(issues, "Business type mismatch between session data and dynamic state")
			}
		}
	}

	return issues
}
