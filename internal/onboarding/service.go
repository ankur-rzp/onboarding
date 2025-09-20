package onboarding

import (
	"context"
	"fmt"
	"strings"
	"time"

	"onboarding-system/internal/config"
	"onboarding-system/internal/storage"

	"github.com/sirupsen/logrus"
)

// Service handles onboarding operations
type Service struct {
	storage storage.Storage
	engine  *Engine
	config  *config.Config
	logger  *logrus.Logger
}

// NewService creates a new onboarding service
func NewService(storage storage.Storage, config *config.Config) *Service {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Service{
		storage: storage,
		engine:  NewEngine(logger),
		config:  config,
		logger:  logger,
	}
}

// StartSession starts a new onboarding session
func (s *Service) StartSession(ctx context.Context, userID, graphID string) (*Session, error) {
	// Get the graph
	graph, err := s.storage.GetGraph(ctx, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Create new session
	session := NewSession(userID, graphID)
	session.CurrentNodeID = graph.StartNodeID

	// Save session
	if err := s.storage.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"user_id":    userID,
		"graph_id":   graphID,
	}).Info("Started new onboarding session")

	return session, nil
}

// GetCurrentNode returns the current node for a session
func (s *Service) GetCurrentNode(ctx context.Context, sessionID string) (*Node, error) {
	session, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	graph, err := s.storage.GetGraph(ctx, session.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	node, exists := graph.Nodes[session.CurrentNodeID]
	if !exists {
		return nil, fmt.Errorf("current node not found: %s", session.CurrentNodeID)
	}

	return node, nil
}

// SubmitNodeData submits data for the current node and moves to the next node
func (s *Service) SubmitNodeData(ctx context.Context, sessionID string, data map[string]interface{}) (*NextStepResult, error) {
	session, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	graph, err := s.storage.GetGraph(ctx, session.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	currentNode, exists := graph.Nodes[session.CurrentNodeID]
	if !exists {
		return nil, fmt.Errorf("current node not found: %s", session.CurrentNodeID)
	}

	// Validate the data against accumulated session data
	// Create a copy of session data and merge with current node data for validation
	validationData := make(map[string]interface{})
	for k, v := range session.Data {
		validationData[k] = v
	}
	for k, v := range data {
		validationData[k] = v
	}

	validationResult := s.engine.ValidateNode(ctx, currentNode, validationData)
	if !validationResult.Valid {
		s.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"node_id":    session.CurrentNodeID,
			"errors":     validationResult.Errors,
		}).Warn("Node validation failed")
		return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
	}

	// Only validate path completeness if we're trying to reach the end node
	// This allows flexible navigation while ensuring all required data is collected before completion

	// Update session data
	for key, value := range data {
		session.Data[key] = value
	}

	// Add to history
	step := SessionStep{
		ID:        fmt.Sprintf("%s-%d", sessionID, len(session.History)),
		NodeID:    session.CurrentNodeID,
		Data:      data,
		Timestamp: time.Now(),
		Action:    "forward",
	}
	session.History = append(session.History, step)

	// First, check if all required nodes are completed - if so, mark as completed
	pathValid, missingNodes := s.engine.ValidatePathCompleteness(ctx, graph, session.CurrentNodeID, session.Data, session.History)

	s.logger.WithFields(logrus.Fields{
		"session_id":    sessionID,
		"current_node":  session.CurrentNodeID,
		"path_valid":    pathValid,
		"missing_nodes": missingNodes,
		"session_data":  session.Data,
	}).Info("Path completeness validation result")

	if pathValid {
		// All required nodes completed - mark as completed
		s.logger.WithFields(logrus.Fields{
			"session_id":   sessionID,
			"current_node": session.CurrentNodeID,
		}).Info("All required nodes completed, marking session as completed")
		session.Status = SessionStatusCompleted
		now := time.Now()
		session.CompletedAt = &now
		// Keep current node ID for UI to show completion
	} else {
		// Not all required nodes completed - find next unfilled node to navigate to
		nextUnfilledNodeID := s.engine.GetFirstMissingNode(ctx, graph, missingNodes, session.Data)
		if nextUnfilledNodeID != "" {
			s.logger.WithFields(logrus.Fields{
				"session_id":    sessionID,
				"current_node":  session.CurrentNodeID,
				"next_node":     nextUnfilledNodeID,
				"missing_nodes": missingNodes,
			}).Info("Navigating to next unfilled required node")
			session.CurrentNodeID = nextUnfilledNodeID
		} else {
			// No specific missing node found, try to get next available node
			nextNodes := s.engine.GetNextNodes(ctx, graph, session.CurrentNodeID, session.Data)
			if len(nextNodes) > 0 {
				// Find the first unfilled node from available next nodes
				nextUnfilledNode := s.findNextUnfilledNode(ctx, graph, nextNodes, session.Data)
				if nextUnfilledNode != nil {
					s.logger.WithFields(logrus.Fields{
						"session_id":     sessionID,
						"current_node":   session.CurrentNodeID,
						"next_node":      nextUnfilledNode.ID,
						"next_node_name": nextUnfilledNode.Name,
					}).Info("Navigating to next unfilled node from available paths")
					session.CurrentNodeID = nextUnfilledNode.ID
				} else {
					// All next nodes are filled, navigate to the first one
					session.CurrentNodeID = nextNodes[0].ID
				}
			} else {
				// No next nodes available, stay on current node
				s.logger.WithFields(logrus.Fields{
					"session_id":    sessionID,
					"current_node":  session.CurrentNodeID,
					"missing_nodes": missingNodes,
				}).Warn("No next nodes available, staying on current node")
			}
		}
	}

	session.UpdatedAt = time.Now()

	// Save updated session
	if err := s.storage.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Prepare result
	result := &NextStepResult{
		NextNodeID:     session.CurrentNodeID,
		AvailablePaths: make([]string, 0),
		CanGoBack:      s.engine.CanGoBack(ctx, graph, currentNode.ID),
		Metadata: map[string]interface{}{
			"validation_warnings": validationResult.Warnings,
			"session_status":      session.Status,
		},
	}

	// If not completed, get available paths
	if session.Status != SessionStatusCompleted {
		nextNodes := s.engine.GetNextNodes(ctx, graph, session.CurrentNodeID, session.Data)
		result.AvailablePaths = make([]string, len(nextNodes))
		for i, node := range nextNodes {
			result.AvailablePaths[i] = node.ID
		}
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":      sessionID,
		"current_node":    currentNode.ID,
		"next_node":       session.CurrentNodeID,
		"available_paths": len(result.AvailablePaths),
		"session_status":  session.Status,
	}).Info("Node data submitted successfully")

	return result, nil
}

// GoBack moves the session back to the previous node
func (s *Service) GoBack(ctx context.Context, sessionID string) (*Node, error) {
	session, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	graph, err := s.storage.GetGraph(ctx, session.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Get previous nodes
	previousNodes := s.engine.GetPreviousNodes(ctx, graph, session.CurrentNodeID)
	if len(previousNodes) == 0 {
		return nil, fmt.Errorf("cannot go back from current node")
	}

	// For simplicity, go back to the most recent previous node
	// In a more sophisticated system, you might want to track the path taken
	previousNode := previousNodes[0]

	// Update session
	session.CurrentNodeID = previousNode.ID
	session.UpdatedAt = time.Now()

	// Add to history
	step := SessionStep{
		ID:        fmt.Sprintf("%s-back-%d", sessionID, len(session.History)),
		NodeID:    previousNode.ID,
		Data:      make(map[string]interface{}),
		Timestamp: time.Now(),
		Action:    "backward",
	}
	session.History = append(session.History, step)

	// Save updated session
	if err := s.storage.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":    sessionID,
		"previous_node": previousNode.ID,
	}).Info("Moved back to previous node")

	return previousNode, nil
}

// GetSession returns a session by ID
func (s *Service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return s.storage.GetSession(ctx, sessionID)
}

func (s *Service) UpdateSession(ctx context.Context, session *Session) error {
	return s.storage.UpdateSession(ctx, session)
}

// GetSessionHistory returns the history of a session
func (s *Service) GetSessionHistory(ctx context.Context, sessionID string) ([]SessionStep, error) {
	session, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session.History, nil
}

// RetrySession retries a failed session
func (s *Service) RetrySession(ctx context.Context, sessionID string) error {
	session, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.Status != SessionStatusFailed {
		return fmt.Errorf("session is not in failed status")
	}

	if session.RetryCount >= s.config.Onboarding.MaxRetries {
		return fmt.Errorf("maximum retry count exceeded")
	}

	// Reset session status
	session.Status = SessionStatusActive
	session.RetryCount++
	session.UpdatedAt = time.Now()

	// Save updated session
	if err := s.storage.SaveSession(ctx, session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":  sessionID,
		"retry_count": session.RetryCount,
	}).Info("Session retry initiated")

	return nil
}

// CreateGraph creates a new onboarding graph
func (s *Service) CreateGraph(ctx context.Context, graph *Graph) error {
	if err := s.storage.SaveGraph(ctx, graph); err != nil {
		return fmt.Errorf("failed to save graph: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"graph_id": graph.ID,
		"name":     graph.Name,
	}).Info("Created new onboarding graph")

	return nil
}

// GetGraph returns a graph by ID
func (s *Service) GetGraph(ctx context.Context, graphID string) (*Graph, error) {
	return s.storage.GetGraph(ctx, graphID)
}

// ListGraphs returns all available graphs
func (s *Service) ListGraphs(ctx context.Context) ([]*Graph, error) {
	return s.storage.ListGraphs(ctx)
}

// ListAllSessions returns all sessions for admin dashboard
func (s *Service) ListAllSessions(ctx context.Context) ([]*Session, error) {
	return s.storage.ListAllSessions(ctx)
}

// ValidatePathCompleteness checks if all required nodes have been completed
func (s *Service) ValidatePathCompleteness(ctx context.Context, graph *Graph, currentNodeID string, sessionData map[string]interface{}, sessionHistory []SessionStep) (bool, []string) {
	return s.engine.ValidatePathCompleteness(ctx, graph, currentNodeID, sessionData, sessionHistory)
}

// findEndNode finds the end node in the graph
func (s *Service) findEndNode(graph *Graph) *Node {
	for _, node := range graph.Nodes {
		if node.Type == "end" {
			return node
		}
	}
	return nil
}

// checkUploadedFiles checks if files have been uploaded for file fields in a node
func (s *Service) checkUploadedFiles(ctx context.Context, sessionID string, node *Node) map[string]bool {
	uploadedFiles := make(map[string]bool)

	// Check each file field in the node
	for _, field := range node.Fields {
		if field.Type == "file" {
			// For now, we'll assume files are uploaded if the field exists in session data
			// In a real implementation, you'd query the upload service
			uploadedFiles[field.ID] = true // Simplified for now
		}
	}

	return uploadedFiles
}

// GetEligibleNodes returns the list of nodes that are eligible for navigation based on current session state
func (s *Service) GetEligibleNodes(ctx context.Context, sessionID string) ([]string, error) {
	session, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	graph, err := s.storage.GetGraph(ctx, session.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	eligibleNodes := make([]string, 0)

	// If no user type selected, only User Type Selection is eligible
	if session.Data == nil || session.Data["user_type"] == nil {
		// Find User Type Selection node
		for nodeID, node := range graph.Nodes {
			if node.Name == "User Type Selection" {
				eligibleNodes = append(eligibleNodes, nodeID)
				break
			}
		}
		return eligibleNodes, nil
	}

	userType := session.Data["user_type"].(string)

	// Based on user type, determine eligible nodes
	for nodeID, node := range graph.Nodes {
		// Skip completion/end nodes
		if node.Type == "end" || node.Name == "Onboarding Complete" {
			continue
		}

		// User Type Selection should remain available even after completion
		// so users can go back and change their selection

		// For individual users, exclude company-specific nodes
		if userType == "individual" {
			if node.Name == "Business Type" || node.Name == "Company Information" || node.Name == "Tax Information" {
				continue
			}
		}

		// For company users, exclude individual-specific nodes
		if userType == "company" {
			if node.Name == "Personal Information" {
				continue
			}
		}

		// All other nodes are eligible
		eligibleNodes = append(eligibleNodes, nodeID)
	}

	return eligibleNodes, nil
}

// findNextUnfilledNode finds the next node that needs to be filled from the available next nodes
func (s *Service) findNextUnfilledNode(ctx context.Context, graph *Graph, nextNodes []*Node, sessionData map[string]interface{}) *Node {
	// Check each next node to see if it has all required fields filled
	for _, node := range nextNodes {
		// Skip completion/end nodes
		if node.Type == "end" || node.Name == "Onboarding Complete" {
			continue
		}

		// Check if this node has all required fields filled
		allRequiredFieldsFilled := true
		for _, field := range node.Fields {
			if field.Required {
				value, exists := sessionData[field.ID]
				if !exists || value == nil || value == "" {
					allRequiredFieldsFilled = false
					break
				}

				// Check if it's not just empty string or whitespace
				if strValue, ok := value.(string); ok {
					if strings.TrimSpace(strValue) == "" {
						allRequiredFieldsFilled = false
						break
					}
				}
			}
		}

		// If this node is not fully filled, it's a candidate
		if !allRequiredFieldsFilled {
			s.logger.WithFields(logrus.Fields{
				"node_id":   node.ID,
				"node_name": node.Name,
			}).Debug("Found unfilled node")
			return node
		}
	}

	// All nodes are filled, return nil
	return nil
}
