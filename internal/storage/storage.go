package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"onboarding-system/internal/config"
	"onboarding-system/internal/types"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Storage interface defines the storage operations
type Storage interface {
	// Session operations
	SaveSession(ctx context.Context, session *types.Session) error
	GetSession(ctx context.Context, sessionID string) (*types.Session, error)
	UpdateSession(ctx context.Context, session *types.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	ListSessions(ctx context.Context, userID string) ([]*types.Session, error)

	// Graph operations
	SaveGraph(ctx context.Context, graph *types.Graph) error
	GetGraph(ctx context.Context, graphID string) (*types.Graph, error)
	UpdateGraph(ctx context.Context, graph *types.Graph) error
	DeleteGraph(ctx context.Context, graphID string) error
	ListGraphs(ctx context.Context) ([]*types.Graph, error)

	// Close closes the storage connections
	Close() error
}

// PostgresRedisStorage implements Storage using PostgreSQL and Redis
type PostgresRedisStorage struct {
	db     *sql.DB
	redis  *redis.Client
	logger *logrus.Logger
}

// New creates a new storage instance
func New(config *config.Config) (Storage, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Check if database configuration is provided
	if !isDatabaseConfigured(config.Database) {
		logger.Info("No database configuration provided, using in-memory storage")
		return NewMemoryStorage(logger), nil
	}

	// Connect to PostgreSQL
	db, err := connectPostgres(config.Database)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect to PostgreSQL, falling back to in-memory storage")
		return NewMemoryStorage(logger), nil
	}

	// Connect to Redis
	redisClient := connectRedis(config.Redis)

	storage := &PostgresRedisStorage{
		db:     db,
		redis:  redisClient,
		logger: logger,
	}

	// Initialize database schema
	if err := storage.initSchema(); err != nil {
		logger.WithError(err).Warn("Failed to initialize database schema, falling back to in-memory storage")
		return NewMemoryStorage(logger), nil
	}

	logger.Info("Using PostgreSQL + Redis storage")
	return storage, nil
}

// isDatabaseConfigured checks if database configuration is provided
func isDatabaseConfigured(dbConfig config.DatabaseConfig) bool {
	// Check if any database configuration is explicitly set
	// If host is empty or all values are defaults, assume no configuration provided
	return dbConfig.Host != "" &&
		dbConfig.User != "" &&
		dbConfig.Password != "" &&
		dbConfig.DBName != ""
}

// connectPostgres connects to PostgreSQL database
func connectPostgres(config config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// connectRedis connects to Redis
func connectRedis(config config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})
}

// initSchema initializes the database schema
func (s *PostgresRedisStorage) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS graphs (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			version VARCHAR(50),
			start_node_id VARCHAR(36),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS nodes (
			id VARCHAR(36) PRIMARY KEY,
			graph_id VARCHAR(36) REFERENCES graphs(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			fields JSONB,
			validation JSONB,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS edges (
			id VARCHAR(36) PRIMARY KEY,
			graph_id VARCHAR(36) REFERENCES graphs(id) ON DELETE CASCADE,
			from_node_id VARCHAR(36) REFERENCES nodes(id) ON DELETE CASCADE,
			to_node_id VARCHAR(36) REFERENCES nodes(id) ON DELETE CASCADE,
			condition JSONB,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			graph_id VARCHAR(36) REFERENCES graphs(id),
			current_node_id VARCHAR(36),
			data JSONB,
			history JSONB,
			status VARCHAR(50) NOT NULL,
			retry_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_graph_id ON nodes(graph_id)`,
		`CREATE INDEX IF NOT EXISTS idx_edges_graph_id ON edges(graph_id)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// SaveSession saves a session to the database
func (s *PostgresRedisStorage) SaveSession(ctx context.Context, session *types.Session) error {
	dataJSON, err := json.Marshal(session.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	historyJSON, err := json.Marshal(session.History)
	if err != nil {
		return fmt.Errorf("failed to marshal session history: %w", err)
	}

	query := `INSERT INTO sessions (id, user_id, graph_id, current_node_id, data, history, status, retry_count, created_at, updated_at, completed_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			  ON CONFLICT (id) DO UPDATE SET
			  current_node_id = EXCLUDED.current_node_id,
			  data = EXCLUDED.data,
			  history = EXCLUDED.history,
			  status = EXCLUDED.status,
			  retry_count = EXCLUDED.retry_count,
			  updated_at = EXCLUDED.updated_at,
			  completed_at = EXCLUDED.completed_at`

	_, err = s.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.GraphID, session.CurrentNodeID,
		dataJSON, historyJSON, session.Status, session.RetryCount,
		session.CreatedAt, session.UpdatedAt, session.CompletedAt)

	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Cache in Redis
	s.cacheSession(ctx, session)

	return nil
}

// GetSession retrieves a session from the database
func (s *PostgresRedisStorage) GetSession(ctx context.Context, sessionID string) (*types.Session, error) {
	// Try to get from cache first
	if session := s.getCachedSession(ctx, sessionID); session != nil {
		return session, nil
	}

	query := `SELECT id, user_id, graph_id, current_node_id, data, history, status, retry_count, created_at, updated_at, completed_at
			  FROM sessions WHERE id = $1`

	var session types.Session
	var dataJSON, historyJSON []byte
	var completedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.GraphID, &session.CurrentNodeID,
		&dataJSON, &historyJSON, &session.Status, &session.RetryCount,
		&session.CreatedAt, &session.UpdatedAt, &completedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(dataJSON, &session.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	if err := json.Unmarshal(historyJSON, &session.History); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session history: %w", err)
	}

	if completedAt.Valid {
		session.CompletedAt = &completedAt.Time
	}

	// Cache the session
	s.cacheSession(ctx, &session)

	return &session, nil
}

// UpdateSession updates a session in the database
func (s *PostgresRedisStorage) UpdateSession(ctx context.Context, session *types.Session) error {
	return s.SaveSession(ctx, session)
}

// DeleteSession deletes a session from the database
func (s *PostgresRedisStorage) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from cache
	s.redis.Del(ctx, fmt.Sprintf("session:%s", sessionID))

	return nil
}

// ListSessions lists sessions for a user
func (s *PostgresRedisStorage) ListSessions(ctx context.Context, userID string) ([]*types.Session, error) {
	query := `SELECT id, user_id, graph_id, current_node_id, data, history, status, retry_count, created_at, updated_at, completed_at
			  FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*types.Session
	for rows.Next() {
		var session types.Session
		var dataJSON, historyJSON []byte
		var completedAt sql.NullTime

		err := rows.Scan(
			&session.ID, &session.UserID, &session.GraphID, &session.CurrentNodeID,
			&dataJSON, &historyJSON, &session.Status, &session.RetryCount,
			&session.CreatedAt, &session.UpdatedAt, &completedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(dataJSON, &session.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
		}

		if err := json.Unmarshal(historyJSON, &session.History); err != nil {
			return nil, fmt.Errorf("failed to unmarshal session history: %w", err)
		}

		if completedAt.Valid {
			session.CompletedAt = &completedAt.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// SaveGraph saves a graph to the database
func (s *PostgresRedisStorage) SaveGraph(ctx context.Context, graph *types.Graph) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Save graph
	graphQuery := `INSERT INTO graphs (id, name, description, version, start_node_id, metadata, created_at, updated_at)
				   VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				   ON CONFLICT (id) DO UPDATE SET
				   name = EXCLUDED.name,
				   description = EXCLUDED.description,
				   version = EXCLUDED.version,
				   start_node_id = EXCLUDED.start_node_id,
				   metadata = EXCLUDED.metadata,
				   updated_at = EXCLUDED.updated_at`

	metadataJSON, err := json.Marshal(graph.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal graph metadata: %w", err)
	}

	_, err = tx.ExecContext(ctx, graphQuery,
		graph.ID, graph.Name, graph.Description, graph.Version,
		graph.StartNodeID, metadataJSON, graph.CreatedAt, graph.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save graph: %w", err)
	}

	// Save nodes
	for _, node := range graph.Nodes {
		if err := s.saveNode(ctx, tx, graph.ID, node); err != nil {
			return fmt.Errorf("failed to save node: %w", err)
		}
	}

	// Save edges
	for _, edge := range graph.Edges {
		if err := s.saveEdge(ctx, tx, graph.ID, edge); err != nil {
			return fmt.Errorf("failed to save edge: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Cache in Redis
	s.cacheGraph(ctx, graph)

	return nil
}

// saveNode saves a node to the database
func (s *PostgresRedisStorage) saveNode(ctx context.Context, tx *sql.Tx, graphID string, node *types.Node) error {
	fieldsJSON, err := json.Marshal(node.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal node fields: %w", err)
	}

	validationJSON, err := json.Marshal(node.Validation)
	if err != nil {
		return fmt.Errorf("failed to marshal node validation: %w", err)
	}

	metadataJSON, err := json.Marshal(node.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal node metadata: %w", err)
	}

	query := `INSERT INTO nodes (id, graph_id, type, name, description, fields, validation, metadata, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			  ON CONFLICT (id) DO UPDATE SET
			  type = EXCLUDED.type,
			  name = EXCLUDED.name,
			  description = EXCLUDED.description,
			  fields = EXCLUDED.fields,
			  validation = EXCLUDED.validation,
			  metadata = EXCLUDED.metadata,
			  updated_at = EXCLUDED.updated_at`

	_, err = tx.ExecContext(ctx, query,
		node.ID, graphID, node.Type, node.Name, node.Description,
		fieldsJSON, validationJSON, metadataJSON, node.CreatedAt, node.UpdatedAt)

	return err
}

// saveEdge saves an edge to the database
func (s *PostgresRedisStorage) saveEdge(ctx context.Context, tx *sql.Tx, graphID string, edge *types.Edge) error {
	conditionJSON, err := json.Marshal(edge.Condition)
	if err != nil {
		return fmt.Errorf("failed to marshal edge condition: %w", err)
	}

	metadataJSON, err := json.Marshal(edge.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal edge metadata: %w", err)
	}

	query := `INSERT INTO edges (id, graph_id, from_node_id, to_node_id, condition, metadata, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)
			  ON CONFLICT (id) DO UPDATE SET
			  from_node_id = EXCLUDED.from_node_id,
			  to_node_id = EXCLUDED.to_node_id,
			  condition = EXCLUDED.condition,
			  metadata = EXCLUDED.metadata`

	_, err = tx.ExecContext(ctx, query,
		edge.ID, graphID, edge.FromNodeID, edge.ToNodeID,
		conditionJSON, metadataJSON, edge.CreatedAt)

	return err
}

// GetGraph retrieves a graph from the database
func (s *PostgresRedisStorage) GetGraph(ctx context.Context, graphID string) (*types.Graph, error) {
	// Try to get from cache first
	if graph := s.getCachedGraph(ctx, graphID); graph != nil {
		return graph, nil
	}

	// Get graph
	graphQuery := `SELECT id, name, description, version, start_node_id, metadata, created_at, updated_at
				   FROM graphs WHERE id = $1`

	var graph types.Graph
	var metadataJSON []byte

	err := s.db.QueryRowContext(ctx, graphQuery, graphID).Scan(
		&graph.ID, &graph.Name, &graph.Description, &graph.Version,
		&graph.StartNodeID, &metadataJSON, &graph.CreatedAt, &graph.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("graph not found")
		}
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataJSON, &graph.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graph metadata: %w", err)
	}

	// Get nodes
	nodes, err := s.getGraphNodes(ctx, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph nodes: %w", err)
	}
	graph.Nodes = nodes

	// Get edges
	edges, err := s.getGraphEdges(ctx, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph edges: %w", err)
	}
	graph.Edges = edges

	// Cache the graph
	s.cacheGraph(ctx, &graph)

	return &graph, nil
}

// getGraphNodes retrieves all nodes for a graph
func (s *PostgresRedisStorage) getGraphNodes(ctx context.Context, graphID string) (map[string]*types.Node, error) {
	query := `SELECT id, type, name, description, fields, validation, metadata, created_at, updated_at
			  FROM nodes WHERE graph_id = $1`

	rows, err := s.db.QueryContext(ctx, query, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	nodes := make(map[string]*types.Node)
	for rows.Next() {
		var node types.Node
		var fieldsJSON, validationJSON, metadataJSON []byte

		err := rows.Scan(
			&node.ID, &node.Type, &node.Name, &node.Description,
			&fieldsJSON, &validationJSON, &metadataJSON,
			&node.CreatedAt, &node.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(fieldsJSON, &node.Fields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node fields: %w", err)
		}

		if err := json.Unmarshal(validationJSON, &node.Validation); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node validation: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &node.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node metadata: %w", err)
		}

		nodes[node.ID] = &node
	}

	return nodes, nil
}

// getGraphEdges retrieves all edges for a graph
func (s *PostgresRedisStorage) getGraphEdges(ctx context.Context, graphID string) (map[string]*types.Edge, error) {
	query := `SELECT id, from_node_id, to_node_id, condition, metadata, created_at
			  FROM edges WHERE graph_id = $1`

	rows, err := s.db.QueryContext(ctx, query, graphID)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}
	defer rows.Close()

	edges := make(map[string]*types.Edge)
	for rows.Next() {
		var edge types.Edge
		var conditionJSON, metadataJSON []byte

		err := rows.Scan(
			&edge.ID, &edge.FromNodeID, &edge.ToNodeID,
			&conditionJSON, &metadataJSON, &edge.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(conditionJSON, &edge.Condition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal edge condition: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &edge.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal edge metadata: %w", err)
		}

		edges[edge.ID] = &edge
	}

	return edges, nil
}

// UpdateGraph updates a graph in the database
func (s *PostgresRedisStorage) UpdateGraph(ctx context.Context, graph *types.Graph) error {
	return s.SaveGraph(ctx, graph)
}

// DeleteGraph deletes a graph from the database
func (s *PostgresRedisStorage) DeleteGraph(ctx context.Context, graphID string) error {
	query := `DELETE FROM graphs WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, graphID)
	if err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	// Remove from cache
	s.redis.Del(ctx, fmt.Sprintf("graph:%s", graphID))

	return nil
}

// ListGraphs lists all graphs
func (s *PostgresRedisStorage) ListGraphs(ctx context.Context) ([]*types.Graph, error) {
	query := `SELECT id, name, description, version, start_node_id, metadata, created_at, updated_at
			  FROM graphs ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list graphs: %w", err)
	}
	defer rows.Close()

	var graphs []*types.Graph
	for rows.Next() {
		var graph types.Graph
		var metadataJSON []byte

		err := rows.Scan(
			&graph.ID, &graph.Name, &graph.Description, &graph.Version,
			&graph.StartNodeID, &metadataJSON, &graph.CreatedAt, &graph.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan graph: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataJSON, &graph.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal graph metadata: %w", err)
		}

		graphs = append(graphs, &graph)
	}

	return graphs, nil
}

// Close closes the storage connections
func (s *PostgresRedisStorage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	if err := s.redis.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %w", err)
	}

	return nil
}

// Cache operations
func (s *PostgresRedisStorage) cacheSession(ctx context.Context, session *types.Session) {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal session for cache")
		return
	}

	key := fmt.Sprintf("session:%s", session.ID)
	s.redis.Set(ctx, key, sessionJSON, 24*time.Hour)
}

func (s *PostgresRedisStorage) getCachedSession(ctx context.Context, sessionID string) *types.Session {
	key := fmt.Sprintf("session:%s", sessionID)
	sessionJSON, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var session types.Session
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal cached session")
		return nil
	}

	return &session
}

func (s *PostgresRedisStorage) cacheGraph(ctx context.Context, graph *types.Graph) {
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal graph for cache")
		return
	}

	key := fmt.Sprintf("graph:%s", graph.ID)
	s.redis.Set(ctx, key, graphJSON, 24*time.Hour)
}

func (s *PostgresRedisStorage) getCachedGraph(ctx context.Context, graphID string) *types.Graph {
	key := fmt.Sprintf("graph:%s", graphID)
	graphJSON, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var graph types.Graph
	if err := json.Unmarshal([]byte(graphJSON), &graph); err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal cached graph")
		return nil
	}

	return &graph
}
