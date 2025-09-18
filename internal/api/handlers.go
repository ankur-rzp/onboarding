package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"onboarding-system/internal/onboarding"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	onboardingService *onboarding.Service
	logger            *logrus.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(onboardingService *onboarding.Service) *Handlers {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Handlers{
		onboardingService: onboardingService,
		logger:            logger,
	}
}

// Router returns the HTTP router with all routes configured
func (h *Handlers) Router() *mux.Router {
	router := mux.NewRouter()

	// Add CORS middleware
	router.Use(h.corsMiddleware)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Graph routes
	api.HandleFunc("/graphs", h.ListGraphs).Methods("GET")
	api.HandleFunc("/graphs", h.CreateGraph).Methods("POST")
	api.HandleFunc("/graphs", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/graphs/{id}", h.GetGraph).Methods("GET")
	api.HandleFunc("/graphs/{id}", h.UpdateGraph).Methods("PUT")
	api.HandleFunc("/graphs/{id}", h.DeleteGraph).Methods("DELETE")
	api.HandleFunc("/graphs/{id}", h.corsHandler).Methods("OPTIONS")

	// Session routes
	api.HandleFunc("/sessions", h.ListSessions).Methods("GET")
	api.HandleFunc("/sessions", h.StartSession).Methods("POST")
	api.HandleFunc("/sessions", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}", h.GetSession).Methods("GET")
	api.HandleFunc("/sessions/{id}", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/current", h.GetCurrentNode).Methods("GET")
	api.HandleFunc("/sessions/{id}/current", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/submit", h.SubmitNodeData).Methods("POST")
	api.HandleFunc("/sessions/{id}/submit", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/back", h.GoBack).Methods("POST")
	api.HandleFunc("/sessions/{id}/back", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/retry", h.RetrySession).Methods("POST")
	api.HandleFunc("/sessions/{id}/retry", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/history", h.GetSessionHistory).Methods("GET")
	api.HandleFunc("/sessions/{id}/history", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/users/{user_id}/sessions", h.ListUserSessions).Methods("GET")
	api.HandleFunc("/users/{user_id}/sessions", h.corsHandler).Methods("OPTIONS")

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	return router
}

// HealthCheck handles health check requests
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// ListGraphs handles listing all graphs
func (h *Handlers) ListGraphs(w http.ResponseWriter, r *http.Request) {
	graphs, err := h.onboardingService.ListGraphs(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list graphs")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphs)
}

// CreateGraph handles creating a new graph
func (h *Handlers) CreateGraph(w http.ResponseWriter, r *http.Request) {
	var graph onboarding.Graph
	if err := json.NewDecoder(r.Body).Decode(&graph); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.onboardingService.CreateGraph(r.Context(), &graph); err != nil {
		h.logger.WithError(err).Error("Failed to create graph")
		http.Error(w, "Failed to create graph", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(graph)
}

// GetGraph handles getting a graph by ID
func (h *Handlers) GetGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	graph, err := h.onboardingService.GetGraph(r.Context(), graphID)
	if err != nil {
		h.logger.WithError(err).WithField("graph_id", graphID).Error("Failed to get graph")
		http.Error(w, "Graph not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}

// UpdateGraph handles updating a graph
func (h *Handlers) UpdateGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	var graph onboarding.Graph
	if err := json.NewDecoder(r.Body).Decode(&graph); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	graph.ID = graphID

	if err := h.onboardingService.CreateGraph(r.Context(), &graph); err != nil {
		h.logger.WithError(err).WithField("graph_id", graphID).Error("Failed to update graph")
		http.Error(w, "Failed to update graph", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}

// DeleteGraph handles deleting a graph
func (h *Handlers) DeleteGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	// Note: This would need to be implemented in the service layer
	h.logger.WithField("graph_id", graphID).Info("Delete graph requested")
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// StartSession handles starting a new onboarding session
func (h *Handlers) StartSession(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID  string `json:"user_id"`
		GraphID string `json:"graph_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.UserID == "" || request.GraphID == "" {
		http.Error(w, "user_id and graph_id are required", http.StatusBadRequest)
		return
	}

	session, err := h.onboardingService.StartSession(r.Context(), request.UserID, request.GraphID)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  request.UserID,
			"graph_id": request.GraphID,
		}).Error("Failed to start session")
		http.Error(w, "Failed to start session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// GetSession handles getting a session by ID
func (h *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get session")
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// GetCurrentNode handles getting the current node for a session
func (h *Handlers) GetCurrentNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	node, err := h.onboardingService.GetCurrentNode(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get current node")
		http.Error(w, "Failed to get current node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// SubmitNodeData handles submitting data for the current node
func (h *Handlers) SubmitNodeData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result, err := h.onboardingService.SubmitNodeData(r.Context(), sessionID, data)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to submit node data")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GoBack handles going back to the previous node
func (h *Handlers) GoBack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	node, err := h.onboardingService.GoBack(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to go back")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// RetrySession handles retrying a failed session
func (h *Handlers) RetrySession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	if err := h.onboardingService.RetrySession(r.Context(), sessionID); err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to retry session")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "retry initiated"})
}

// GetSessionHistory handles getting the history of a session
func (h *Handlers) GetSessionHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	history, err := h.onboardingService.GetSessionHistory(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get session history")
		http.Error(w, "Failed to get session history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// ListUserSessions handles listing sessions for a user
func (h *Handlers) ListUserSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Note: This would need to be implemented in the service layer with pagination
	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}).Info("List user sessions requested")

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// ListSessions handles listing all sessions
func (h *Handlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Note: This would need to be implemented in the service layer
	// For now, return empty array as placeholder
	h.logger.Info("List sessions requested")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

// corsHandler handles CORS preflight requests
func (h *Handlers) corsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400")

	w.WriteHeader(http.StatusOK)
}

// corsMiddleware handles CORS headers
func (h *Handlers) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}
