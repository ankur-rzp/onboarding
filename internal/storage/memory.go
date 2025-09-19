package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"onboarding-system/internal/types"

	"github.com/sirupsen/logrus"
)

// MemoryStorage implements Storage using in-memory data structures
type MemoryStorage struct {
	graphs   map[string]*types.Graph
	sessions map[string]*types.Session
	mutex    sync.RWMutex
	logger   *logrus.Logger
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage(logger *logrus.Logger) *MemoryStorage {
	return &MemoryStorage{
		graphs:   make(map[string]*types.Graph),
		sessions: make(map[string]*types.Session),
		logger:   logger,
	}
}

// SaveSession saves a session to memory
func (m *MemoryStorage) SaveSession(ctx context.Context, session *types.Session) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session.UpdatedAt = time.Now()
	m.sessions[session.ID] = session

	m.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"user_id":    session.UserID,
		"status":     session.Status,
	}).Debug("Session saved to memory")

	return nil
}

// GetSession retrieves a session from memory
func (m *MemoryStorage) GetSession(ctx context.Context, sessionID string) (*types.Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Return a copy to avoid race conditions
	sessionCopy := *session
	return &sessionCopy, nil
}

// UpdateSession updates a session in memory
func (m *MemoryStorage) UpdateSession(ctx context.Context, session *types.Session) error {
	return m.SaveSession(ctx, session)
}

// DeleteSession deletes a session from memory
func (m *MemoryStorage) DeleteSession(ctx context.Context, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.sessions, sessionID)

	m.logger.WithField("session_id", sessionID).Debug("Session deleted from memory")
	return nil
}

// ListSessions lists sessions for a user
func (m *MemoryStorage) ListSessions(ctx context.Context, userID string) ([]*types.Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var userSessions []*types.Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			// Return a copy to avoid race conditions
			sessionCopy := *session
			userSessions = append(userSessions, &sessionCopy)
		}
	}

	return userSessions, nil
}

// ListAllSessions lists all sessions for admin dashboard
func (m *MemoryStorage) ListAllSessions(ctx context.Context) ([]*types.Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var allSessions []*types.Session
	for _, session := range m.sessions {
		// Return a copy to avoid race conditions
		sessionCopy := *session
		allSessions = append(allSessions, &sessionCopy)
	}

	return allSessions, nil
}

// SaveGraph saves a graph to memory
func (m *MemoryStorage) SaveGraph(ctx context.Context, graph *types.Graph) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	graph.UpdatedAt = time.Now()
	m.graphs[graph.ID] = graph

	m.logger.WithFields(logrus.Fields{
		"graph_id": graph.ID,
		"name":     graph.Name,
	}).Debug("Graph saved to memory")

	return nil
}

// GetGraph retrieves a graph from memory
func (m *MemoryStorage) GetGraph(ctx context.Context, graphID string) (*types.Graph, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	graph, exists := m.graphs[graphID]
	if !exists {
		return nil, fmt.Errorf("graph not found")
	}

	// Return a copy to avoid race conditions
	graphCopy := *graph
	return &graphCopy, nil
}

// UpdateGraph updates a graph in memory
func (m *MemoryStorage) UpdateGraph(ctx context.Context, graph *types.Graph) error {
	return m.SaveGraph(ctx, graph)
}

// DeleteGraph deletes a graph from memory
func (m *MemoryStorage) DeleteGraph(ctx context.Context, graphID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.graphs, graphID)

	m.logger.WithField("graph_id", graphID).Debug("Graph deleted from memory")
	return nil
}

// ListGraphs lists all graphs
func (m *MemoryStorage) ListGraphs(ctx context.Context) ([]*types.Graph, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var graphs []*types.Graph
	for _, graph := range m.graphs {
		// Return a copy to avoid race conditions
		graphCopy := *graph
		graphs = append(graphs, &graphCopy)
	}

	return graphs, nil
}

// Close closes the memory storage (no-op for in-memory)
func (m *MemoryStorage) Close() error {
	m.logger.Info("Memory storage closed")
	return nil
}

// GetStats returns storage statistics
func (m *MemoryStorage) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"graphs_count":   len(m.graphs),
		"sessions_count": len(m.sessions),
		"storage_type":   "memory",
	}
}

// ClearAll clears all data (useful for testing)
func (m *MemoryStorage) ClearAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.graphs = make(map[string]*types.Graph)
	m.sessions = make(map[string]*types.Session)

	m.logger.Info("All data cleared from memory storage")
}
