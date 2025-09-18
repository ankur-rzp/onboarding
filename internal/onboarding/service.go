package onboarding

import (
	"context"
	"fmt"
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

	// Validate the data
	validationResult := s.engine.ValidateNode(ctx, currentNode, data)
	if !validationResult.Valid {
		s.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"node_id":    session.CurrentNodeID,
			"errors":     validationResult.Errors,
		}).Warn("Node validation failed")
		return nil, fmt.Errorf("validation failed: %v", validationResult.Errors)
	}

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

	// Determine next node
	nextNodes := s.engine.GetNextNodes(ctx, graph, session.CurrentNodeID, session.Data)
	if len(nextNodes) == 0 {
		// No next nodes, session is complete
		session.Status = SessionStatusCompleted
		now := time.Now()
		session.CompletedAt = &now
		session.CurrentNodeID = ""
	} else if len(nextNodes) == 1 {
		// Single next node
		session.CurrentNodeID = nextNodes[0].ID
	} else {
		// Multiple next nodes - this is a decision point
		// For now, we'll return the available paths and let the client choose
		// In a more sophisticated system, you might have decision logic here
	}

	session.UpdatedAt = time.Now()

	// Save updated session
	if err := s.storage.SaveSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Prepare result
	result := &NextStepResult{
		NextNodeID:     session.CurrentNodeID,
		AvailablePaths: make([]string, len(nextNodes)),
		CanGoBack:      s.engine.CanGoBack(ctx, graph, currentNode.ID),
		Metadata: map[string]interface{}{
			"validation_warnings": validationResult.Warnings,
			"session_status":      session.Status,
		},
	}

	for i, node := range nextNodes {
		result.AvailablePaths[i] = node.ID
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":      sessionID,
		"current_node":    currentNode.ID,
		"next_node":       session.CurrentNodeID,
		"available_paths": len(nextNodes),
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
