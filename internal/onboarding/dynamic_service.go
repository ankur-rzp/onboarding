package onboarding

import (
	"context"
	"fmt"
	"time"

	"onboarding-system/internal/config"
	"onboarding-system/internal/storage"

	"github.com/sirupsen/logrus"
)

// DynamicService extends the regular service with dynamic node management
type DynamicService struct {
	*Service
	dynamicEngine      *DynamicEngine
	dynamicGraphs      map[string]*DynamicGraph // graphID -> DynamicGraph
	persistenceManager *DynamicPersistenceManager
	logger             *logrus.Logger
}

// NewDynamicService creates a new dynamic service
func NewDynamicService(storage storage.Storage, config *config.Config, logger *logrus.Logger) *DynamicService {
	return &DynamicService{
		Service:            NewService(storage, config),
		dynamicEngine:      NewDynamicEngine(logger),
		dynamicGraphs:      make(map[string]*DynamicGraph),
		persistenceManager: NewDynamicPersistenceManager(logger),
		logger:             logger,
	}
}

// StartDynamicSession starts a session with dynamic node management
func (ds *DynamicService) StartDynamicSession(ctx context.Context, graphID, userID string) (*Session, error) {
	// Start regular session
	session, err := ds.Service.StartSession(ctx, userID, graphID)
	if err != nil {
		return nil, err
	}

	// Get the graph using the base service
	graph, err := ds.Service.GetGraph(ctx, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Determine business type from session data or use default
	businessType := "individual" // default
	if session.Data != nil {
		if bt, exists := session.Data["business_type"]; exists {
			businessType = fmt.Sprintf("%v", bt)
		}
	}

	// Convert to dynamic graph if not already done
	if _, exists := ds.dynamicGraphs[graphID]; !exists {
		ds.dynamicGraphs[graphID] = ds.dynamicEngine.ConvertToDynamicGraph(graph, businessType)
	}

	dynamicGraph := ds.dynamicGraphs[graphID]

	// Restore dynamic state if it exists
	if err := ds.persistenceManager.RestoreDynamicState(session, dynamicGraph); err != nil {
		ds.logger.WithFields(logrus.Fields{
			"session_id": session.ID,
			"error":      err,
		}).Warn("Failed to restore dynamic state, using initial state")
	}

	// Save initial dynamic state
	ds.persistenceManager.SaveDynamicState(session, dynamicGraph, businessType)

	// Save session with dynamic state
	if err := ds.Service.SaveSession(ctx, session); err != nil {
		ds.logger.WithFields(logrus.Fields{
			"session_id": session.ID,
			"error":      err,
		}).Error("Failed to save session with dynamic state")
	}

	ds.logger.WithFields(logrus.Fields{
		"session_id":    session.ID,
		"graph_id":      graphID,
		"user_id":       userID,
		"business_type": businessType,
	}).Info("Started dynamic session with persistence")

	return session, nil
}

// SubmitNodeDataDynamic submits node data and updates dynamic node status
func (ds *DynamicService) SubmitNodeDataDynamic(ctx context.Context, sessionID string, data map[string]interface{}) (*NextStepResult, error) {
	// Get session
	session, err := ds.Service.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Get dynamic graph
	dynamicGraph, exists := ds.dynamicGraphs[session.GraphID]
	if !exists {
		// Fallback to regular service
		return ds.Service.SubmitNodeData(ctx, sessionID, data)
	}

	// Get current node from the graph
	currentNode, exists := dynamicGraph.Graph.Nodes[session.CurrentNodeID]
	if !exists {
		return nil, fmt.Errorf("current node not found: %s", session.CurrentNodeID)
	}

	// Validate node data using the base engine
	validationResult := ds.dynamicEngine.ValidateNode(ctx, currentNode, data)
	if !validationResult.Valid {
		return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
	}

	// Update session data
	if session.Data == nil {
		session.Data = make(map[string]interface{})
	}
	for k, v := range data {
		session.Data[k] = v
	}

	// Mark current node as completed in dynamic graph
	dynamicGraph.OnNodeCompleted(session.CurrentNodeID, session.Data)

	// Notify observers of data changes
	for fieldID, value := range data {
		dynamicGraph.OnNodeDataChanged(session.CurrentNodeID, fieldID, value, session.Data)
	}

	// Determine business type for persistence
	businessType := "individual" // default
	if session.Data != nil {
		if bt, exists := session.Data["business_type"]; exists {
			businessType = fmt.Sprintf("%v", bt)
		}
	}

	// Save dynamic state after changes
	ds.persistenceManager.SaveDynamicState(session, dynamicGraph, businessType)

	// Determine next node
	nextNode := ds.determineNextNodeDynamic(dynamicGraph, session)
	if nextNode != nil {
		session.CurrentNodeID = nextNode.ID
		session.UpdatedAt = time.Now()
	} else {
		// Check if we can complete
		completionStatus := dynamicGraph.GetCompletionStatus()
		if canComplete, ok := completionStatus["can_complete"].(bool); ok && canComplete {
			// Find the end node
			endNode := ds.findEndNode(dynamicGraph.Graph)
			if endNode != nil {
				session.CurrentNodeID = endNode.ID
				session.Status = SessionStatusCompleted
				now := time.Now()
				session.CompletedAt = &now
			}
		}
	}

	// Save session
	if err := ds.Service.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Prepare result
	result := &NextStepResult{
		NextNodeID:     session.CurrentNodeID,
		AvailablePaths: []string{session.CurrentNodeID},
		CanGoBack:      len(session.History) > 0,
		Metadata: map[string]interface{}{
			"session_status":      string(session.Status),
			"validation_warnings": validationResult.Warnings,
			"dynamic_status":      dynamicGraph.GetCompletionStatus(),
		},
	}

	completionStatus := dynamicGraph.GetCompletionStatus()
	ds.logger.WithFields(logrus.Fields{
		"session_id":        sessionID,
		"next_node_id":      session.CurrentNodeID,
		"completion_status": completionStatus,
	}).Info("Submitted node data with dynamic status update")

	return result, nil
}

// GetDynamicNodeStatus returns the current status of all nodes with persistence
func (ds *DynamicService) GetDynamicNodeStatus(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	// Get session
	session, err := ds.Service.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Get dynamic graph
	dynamicGraph, exists := ds.dynamicGraphs[session.GraphID]
	if !exists {
		return nil, fmt.Errorf("dynamic graph not found for session: %s", sessionID)
	}

	// Restore dynamic state if needed
	if err := ds.persistenceManager.RestoreDynamicState(session, dynamicGraph); err != nil {
		ds.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"error":      err,
		}).Warn("Failed to restore dynamic state")
	}

	// Build response
	response := map[string]interface{}{
		"session_id":        sessionID,
		"business_type":     session.DynamicState.BusinessType,
		"last_evaluated_at": session.DynamicState.LastEvaluatedAt,
		"completion_status": session.DynamicState.CompletionStatus,
		"nodes":             make(map[string]interface{}),
	}

	// Add node statuses
	for nodeID, dynamicNode := range dynamicGraph.DynamicNodes {
		nodeInfo := map[string]interface{}{
			"id":             nodeID,
			"name":           dynamicNode.Name,
			"status":         string(dynamicNode.Status),
			"initial_status": string(dynamicNode.InitialStatus),
			"dependencies":   dynamicNode.Dependencies,
			"type":           dynamicNode.Type,
		}
		response["nodes"].(map[string]interface{})[nodeID] = nodeInfo
	}

	return response, nil
}

// UpdateBusinessTypeDynamic updates the business type and recalculates all node statuses
func (ds *DynamicService) UpdateBusinessTypeDynamic(ctx context.Context, sessionID string, businessType string) error {
	// Get session
	session, err := ds.Service.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get the graph
	graph, err := ds.Service.GetGraph(ctx, session.GraphID)
	if err != nil {
		return fmt.Errorf("failed to get graph: %w", err)
	}

	// Update session data
	if session.Data == nil {
		session.Data = make(map[string]interface{})
	}
	session.Data["business_type"] = businessType

	// Create new dynamic graph with new business type
	dynamicGraph := ds.dynamicEngine.ConvertToDynamicGraph(graph, businessType)
	ds.dynamicGraphs[session.GraphID] = dynamicGraph

	// Save dynamic state
	ds.persistenceManager.SaveDynamicState(session, dynamicGraph, businessType)

	// Save session
	if err := ds.Service.SaveSession(ctx, session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	ds.logger.WithFields(logrus.Fields{
		"session_id":    sessionID,
		"business_type": businessType,
	}).Info("Updated business type and recalculated dynamic state")

	return nil
}

// GetDynamicStateSummary returns a summary of the dynamic state
func (ds *DynamicService) GetDynamicStateSummary(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	// Get session
	session, err := ds.Service.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Get dynamic graph
	dynamicGraph, exists := ds.dynamicGraphs[session.GraphID]
	if !exists {
		return nil, fmt.Errorf("dynamic graph not found for session: %s", sessionID)
	}

	// Validate dynamic state
	issues := ds.persistenceManager.ValidateDynamicState(session, dynamicGraph)
	if len(issues) > 0 {
		ds.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"issues":     issues,
		}).Warn("Dynamic state validation issues found")
	}

	// Get summary
	summary := ds.persistenceManager.GetDynamicStateSummary(session)
	summary["validation_issues"] = issues

	return summary, nil
}

// determineNextNodeDynamic determines the next node based on dynamic status
func (ds *DynamicService) determineNextNodeDynamic(dynamicGraph *DynamicGraph, session *Session) *Node {
	// Priority order: mandatory -> dependent -> optional
	priorities := []NodeStatus{NodeStatusMandatory, NodeStatusDependent, NodeStatusOptional}

	for _, priority := range priorities {
		for nodeID, dynamicNode := range dynamicGraph.DynamicNodes {
			// Skip if not the right status
			if dynamicNode.Status != priority {
				continue
			}

			// Skip if already completed
			if dynamicNode.Status == NodeStatusCompleted {
				continue
			}

			// Skip if disabled
			if dynamicNode.Status == NodeStatusDisabled {
				continue
			}

			// Skip if this is the current node
			if nodeID == session.CurrentNodeID {
				continue
			}

			// Check if this node is accessible from current node
			if ds.isNodeAccessible(dynamicGraph.Graph, session.CurrentNodeID, nodeID, session.Data) {
				return dynamicNode.Node
			}
		}
	}

	return nil
}

// isNodeAccessible checks if a node is accessible from the current node
func (ds *DynamicService) isNodeAccessible(graph *Graph, fromNodeID, toNodeID string, sessionData map[string]interface{}) bool {
	// Check if there's a direct edge
	for _, edge := range graph.Edges {
		if edge.FromNodeID == fromNodeID && edge.ToNodeID == toNodeID {
			// Check edge condition using the base engine
			return ds.dynamicEngine.evaluateEdgeCondition(edge.Condition, sessionData)
		}
	}

	// For dynamic graphs, we allow more flexible navigation
	// Check if the target node is in a reasonable position in the flow
	return true // For now, allow navigation to any node
}

// GetEligibleNodesDynamic returns nodes that are eligible for navigation based on dynamic status
func (ds *DynamicService) GetEligibleNodesDynamic(ctx context.Context, sessionID string) ([]string, error) {
	session, err := ds.Service.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	dynamicGraph, exists := ds.dynamicGraphs[session.GraphID]
	if !exists {
		// Fallback to regular service
		return ds.Service.GetEligibleNodes(ctx, sessionID)
	}

	eligibleNodes := make([]string, 0)

	for nodeID, dynamicNode := range dynamicGraph.DynamicNodes {
		// Skip end/completion nodes
		if dynamicNode.Type == "end" || dynamicNode.Name == "Onboarding Complete" {
			continue
		}

		// Include nodes that are not disabled
		if dynamicNode.Status != NodeStatusDisabled {
			eligibleNodes = append(eligibleNodes, nodeID)
		}
	}

	ds.logger.WithFields(logrus.Fields{
		"session_id":     sessionID,
		"eligible_nodes": len(eligibleNodes),
	}).Info("Retrieved eligible nodes with dynamic status")

	return eligibleNodes, nil
}
