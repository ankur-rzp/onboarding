package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"onboarding-system/internal/onboarding"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// UploadProgress tracks file upload progress
type UploadProgress struct {
	SessionID    string    `json:"session_id"`
	FieldID      string    `json:"field_id"`
	FileName     string    `json:"file_name"`
	Progress     int       `json:"progress"` // 0-100
	Status       string    `json:"status"`   // "uploading", "completed", "failed"
	Error        string    `json:"error,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
	FileSize     int64     `json:"file_size"`
	UploadedSize int64     `json:"uploaded_size"`
}

// Handlers contains all HTTP handlers
type Handlers struct {
	onboardingService *onboarding.Service
	logger            *logrus.Logger
	uploadProgress    map[string]*UploadProgress // sessionID_fieldID -> progress
	uploadDir         string
	mutex             sync.RWMutex
}

// NewHandlers creates a new handlers instance
func NewHandlers(onboardingService *onboarding.Service) *Handlers {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create upload directory
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.WithError(err).Error("Failed to create upload directory")
	}

	return &Handlers{
		onboardingService: onboardingService,
		logger:            logger,
		uploadProgress:    make(map[string]*UploadProgress),
		uploadDir:         uploadDir,
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
	api.HandleFunc("/sessions/{id}/navigate/{node_id}", h.NavigateToNode).Methods("POST")
	api.HandleFunc("/sessions/{id}/navigate/{node_id}", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/submit", h.SubmitNodeData).Methods("POST")
	api.HandleFunc("/sessions/{id}/submit", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/complete", h.CompleteSession).Methods("POST")
	api.HandleFunc("/sessions/{id}/complete", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/back", h.GoBack).Methods("POST")
	api.HandleFunc("/sessions/{id}/back", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/retry", h.RetrySession).Methods("POST")
	api.HandleFunc("/sessions/{id}/retry", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/history", h.GetSessionHistory).Methods("GET")
	api.HandleFunc("/sessions/{id}/history", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/users/{user_id}/sessions", h.ListUserSessions).Methods("GET")
	api.HandleFunc("/users/{user_id}/sessions", h.corsHandler).Methods("OPTIONS")

	// File upload routes
	api.HandleFunc("/sessions/{id}/upload/{field_id}", h.UploadFile).Methods("POST")
	api.HandleFunc("/sessions/{id}/upload/{field_id}", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/upload/{field_id}/progress", h.GetUploadProgress).Methods("GET")
	api.HandleFunc("/sessions/{id}/upload/{field_id}/progress", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/uploads", h.GetSessionUploads).Methods("GET")
	api.HandleFunc("/sessions/{id}/uploads", h.corsHandler).Methods("OPTIONS")

	// Admin routes
	api.HandleFunc("/admin/sessions", h.ListAllSessions).Methods("GET")
	api.HandleFunc("/admin/sessions", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/admin/sessions/{id}/details", h.GetSessionDetails).Methods("GET")
	api.HandleFunc("/admin/sessions/{id}/details", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/admin/graphs/{id}/visual", h.GetGraphVisual).Methods("GET")
	api.HandleFunc("/admin/graphs/{id}/visual", h.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/admin/sessions/{id}/graph-visual", h.GetSessionGraphVisual).Methods("GET")
	api.HandleFunc("/admin/sessions/{id}/graph-visual", h.corsHandler).Methods("OPTIONS")

	// Eligible nodes route
	api.HandleFunc("/sessions/{id}/eligible-nodes", h.GetEligibleNodes).Methods("GET")
	api.HandleFunc("/sessions/{id}/eligible-nodes", h.corsHandler).Methods("OPTIONS")

	// File download route
	api.HandleFunc("/files/{path:.*}", h.DownloadFile).Methods("GET")
	api.HandleFunc("/files/{path:.*}", h.corsHandler).Methods("OPTIONS")

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// Serve HTML files directly
	router.HandleFunc("/", h.ServeIndex).Methods("GET")
	router.HandleFunc("/dynamic-onboarding-ui.html", h.ServeHTML("dynamic-onboarding-ui.html")).Methods("GET")
	router.HandleFunc("/admin-dashboard.html", h.ServeHTML("admin-dashboard.html")).Methods("GET")
	router.HandleFunc("/test-ui.html", h.ServeHTML("test-ui.html")).Methods("GET")

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

// NavigateToNode allows the UI to navigate to a specific node
func (h *Handlers) NavigateToNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	nodeID := vars["node_id"]

	// Check if the node exists and is accessible
	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get session")
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	graph, err := h.onboardingService.GetGraph(r.Context(), session.GraphID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get graph")
		http.Error(w, "Graph not found", http.StatusNotFound)
		return
	}

	node, exists := graph.Nodes[nodeID]
	if !exists {
		h.logger.WithField("node_id", nodeID).Error("Node not found")
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	// Check if the node is accessible from the current session state
	// For now, we'll allow navigation to any node that exists in the graph
	// In a more sophisticated system, you might want to check if the node is reachable

	// Update the session's current node
	session.CurrentNodeID = nodeID
	session.UpdatedAt = time.Now()

	if err := h.onboardingService.UpdateSession(r.Context(), session); err != nil {
		h.logger.WithError(err).Error("Failed to update session")
		http.Error(w, "Failed to navigate to node", http.StatusInternalServerError)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"node_id":    nodeID,
		"node_name":  node.Name,
	}).Info("Navigated to node")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"node":    node,
		"message": fmt.Sprintf("Successfully navigated to %s", node.Name),
	})
}

// SubmitNodeData handles submitting data for the current node
func (h *Handlers) SubmitNodeData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Check for ongoing uploads before allowing data submission
	if h.CheckOngoingUploads(sessionID) {
		h.logger.WithField("session_id", sessionID).Warn("Blocking data submission due to ongoing uploads")
		http.Error(w, "Cannot submit data while uploads are in progress. Please wait for uploads to complete or cancel them.", http.StatusConflict)
		return
	}

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

// CompleteSession handles completing the onboarding session
func (h *Handlers) CompleteSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Check for ongoing uploads before allowing completion
	if h.CheckOngoingUploads(sessionID) {
		h.logger.WithField("session_id", sessionID).Warn("Blocking completion due to ongoing uploads")
		http.Error(w, "Cannot complete onboarding while uploads are in progress. Please wait for uploads to complete or cancel them.", http.StatusConflict)
		return
	}

	// Get the session to check if completion is valid
	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get session")
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get the graph to validate completion
	graph, err := h.onboardingService.GetGraph(r.Context(), session.GraphID)
	if err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to get graph")
		http.Error(w, "Graph not found", http.StatusNotFound)
		return
	}

	// Check if all required nodes have been completed
	pathValid, missingNodes := h.onboardingService.ValidatePathCompleteness(r.Context(), graph, session.CurrentNodeID, session.Data, session.History)
	if !pathValid {
		h.logger.WithFields(logrus.Fields{
			"session_id":    sessionID,
			"current_node":  session.CurrentNodeID,
			"missing_nodes": missingNodes,
		}).Warn("Cannot complete session - missing required nodes")
		http.Error(w, fmt.Sprintf("Cannot complete onboarding yet: you must complete the following steps first: %v", missingNodes), http.StatusBadRequest)
		return
	}

	// All required nodes completed, mark session as complete
	session.Status = "completed"
	now := time.Now()
	session.CompletedAt = &now
	session.CurrentNodeID = ""
	session.UpdatedAt = now

	if err := h.onboardingService.UpdateSession(r.Context(), session); err != nil {
		h.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to update session")
		http.Error(w, "Failed to complete session", http.StatusInternalServerError)
		return
	}

	h.logger.WithField("session_id", sessionID).Info("Session completed successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"message":        "Onboarding completed successfully!",
		"session_status": "completed",
		"completed_at":   now,
	})
}

// GoBack handles going back to the previous node
func (h *Handlers) GoBack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Check for ongoing uploads before allowing navigation
	if h.CheckOngoingUploads(sessionID) {
		h.logger.WithField("session_id", sessionID).Warn("Blocking navigation due to ongoing uploads")
		http.Error(w, "Cannot navigate while uploads are in progress. Please wait for uploads to complete or cancel them.", http.StatusConflict)
		return
	}

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

// UploadFile handles file upload with progress tracking
func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	fieldID := vars["field_id"]

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse multipart form")
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.WithError(err).Error("Failed to get file from form")
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create progress tracking with unique upload ID
	uploadID := fmt.Sprintf("%s_%s_%d", sessionID, fieldID, time.Now().UnixNano())
	progressKey := fmt.Sprintf("%s_%s", sessionID, fieldID)
	progress := &UploadProgress{
		SessionID:    sessionID,
		FieldID:      fieldID,
		FileName:     header.Filename,
		Progress:     0,
		Status:       "uploading",
		UploadedAt:   time.Now(),
		FileSize:     header.Size,
		UploadedSize: 0,
	}
	h.mutex.Lock()
	h.uploadProgress[progressKey] = progress
	h.mutex.Unlock()

	// Create file path
	filePath := filepath.Join(h.uploadDir, sessionID, fieldID, header.Filename)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		h.logger.WithError(err).Error("Failed to create upload directory")
		progress.Status = "failed"
		progress.Error = "Failed to create directory"
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Create destination file
	destFile, err := os.Create(filePath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create destination file")
		progress.Status = "failed"
		progress.Error = "Failed to create file"
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy file with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			_, writeErr := destFile.Write(buffer[:n])
			if writeErr != nil {
				h.logger.WithError(writeErr).Error("Failed to write file")
				progress.Status = "failed"
				progress.Error = "Failed to write file"
				http.Error(w, "Failed to write file", http.StatusInternalServerError)
				return
			}
			progress.UploadedSize += int64(n)
			progress.Progress = int((progress.UploadedSize * 100) / progress.FileSize)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			h.logger.WithError(err).Error("Failed to read file")
			progress.Status = "failed"
			progress.Error = "Failed to read file"
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
	}

	// Start async processing in a goroutine
	go func() {
		// Simulate file processing with deliberate delay for demo purposes
		// In production, this would be actual file processing/validation
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			h.mutex.Lock()
			progress.Progress = 100 + (i * 2) // Simulate progress from 100% to 160%
			h.mutex.Unlock()
		}

		// Mark as completed
		h.mutex.Lock()
		progress.Status = "completed"
		progress.Progress = 100
		h.mutex.Unlock()

		h.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"field_id":   fieldID,
			"file_name":  header.Filename,
			"file_size":  header.Size,
		}).Info("File processing completed")
	}()

	h.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"field_id":   fieldID,
		"file_name":  header.Filename,
		"file_size":  header.Size,
	}).Info("File upload started, processing in background")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"upload_id": uploadID,
		"file_path": filePath,
		"file_name": header.Filename,
		"file_size": header.Size,
		"status":    "uploading",
		"progress":  progress,
	})
}

// GetUploadProgress returns the upload progress for a specific file
func (h *Handlers) GetUploadProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	fieldID := vars["field_id"]

	progressKey := fmt.Sprintf("%s_%s", sessionID, fieldID)
	h.mutex.RLock()
	progress, exists := h.uploadProgress[progressKey]
	h.mutex.RUnlock()
	if !exists {
		http.Error(w, "Upload not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

// GetSessionUploads returns all uploads for a session
func (h *Handlers) GetSessionUploads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	h.mutex.RLock()
	var sessionUploads []*UploadProgress
	for key, progress := range h.uploadProgress {
		if strings.HasPrefix(key, sessionID+"_") {
			sessionUploads = append(sessionUploads, progress)
		}
	}
	h.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessionUploads)
}

// CheckOngoingUploads checks if there are any ongoing uploads for a session
func (h *Handlers) CheckOngoingUploads(sessionID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for key, progress := range h.uploadProgress {
		if strings.HasPrefix(key, sessionID+"_") && progress.Status == "uploading" {
			return true
		}
	}
	return false
}

// ListAllSessions returns all sessions for admin dashboard
func (h *Handlers) ListAllSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.onboardingService.ListAllSessions(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list all sessions")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Enhance sessions with additional info for admin dashboard
	enhancedSessions := make([]map[string]interface{}, 0, len(sessions))
	for _, session := range sessions {
		// Calculate progress
		progress := h.calculateSessionProgress(session)

		enhancedSession := map[string]interface{}{
			"id":              session.ID,
			"user_id":         session.UserID,
			"graph_id":        session.GraphID,
			"current_node_id": session.CurrentNodeID,
			"status":          session.Status,
			"progress":        progress,
			"retry_count":     session.RetryCount,
			"created_at":      session.CreatedAt,
			"updated_at":      session.UpdatedAt,
			"completed_at":    session.CompletedAt,
		}

		// Add current node name if available
		if session.CurrentNodeID != "" {
			// Try to get graph to find node name
			if graph, err := h.onboardingService.GetGraph(r.Context(), session.GraphID); err == nil {
				if node, exists := graph.Nodes[session.CurrentNodeID]; exists {
					enhancedSession["current_node_name"] = node.Name
				}
			}
		}

		enhancedSessions = append(enhancedSessions, enhancedSession)
	}

	h.logger.WithField("count", len(sessions)).Info("List all sessions requested (admin)")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enhancedSessions)
}

// GetSessionDetails returns detailed information about a session for admin
func (h *Handlers) GetSessionDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Get session details
	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get session details")
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get session history
	history, err := h.onboardingService.GetSessionHistory(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get session history")
		http.Error(w, "Failed to get session history", http.StatusInternalServerError)
		return
	}

	// Get current node details
	currentNode, err := h.onboardingService.GetCurrentNode(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get current node")
		http.Error(w, "Failed to get current node", http.StatusInternalServerError)
		return
	}

	// Get graph to understand node structure
	graph, err := h.onboardingService.GetGraph(r.Context(), session.GraphID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get graph")
		http.Error(w, "Failed to get graph", http.StatusInternalServerError)
		return
	}

	// Build comprehensive node data with historical information
	h.logger.Info("Building comprehensive node data")
	nodeData := h.buildComprehensiveNodeData(session, history, graph)
	h.logger.WithField("node_count", len(nodeData)).Info("Built node data")

	// Get uploaded files information
	h.logger.Info("Getting uploaded files")
	uploadedFiles := h.getSessionUploadedFiles(sessionID)
	h.logger.WithField("file_count", len(uploadedFiles)).Info("Got uploaded files")

	response := map[string]interface{}{
		"session":         session,
		"history":         history,
		"current_node":    currentNode,
		"node_data":       nodeData,
		"uploaded_files":  uploadedFiles,
		"upload_progress": h.getSessionUploadProgress(sessionID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetGraphVisual returns visual representation of the graph for admin
func (h *Handlers) GetGraphVisual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	// Get graph
	graph, err := h.onboardingService.GetGraph(r.Context(), graphID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get graph")
		http.Error(w, "Graph not found", http.StatusNotFound)
		return
	}

	// Convert to visual format
	visualGraph := h.convertToVisualGraph(graph)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visualGraph)
}

// GetSessionGraphVisual returns visual representation of the graph with session-specific path highlighting
func (h *Handlers) GetSessionGraphVisual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Get session
	session, err := h.onboardingService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get session")
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get graph
	graph, err := h.onboardingService.GetGraph(r.Context(), session.GraphID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get graph")
		http.Error(w, "Graph not found", http.StatusNotFound)
		return
	}

	// Convert to visual format with session path highlighting
	visualGraph := h.convertToSessionVisualGraph(graph, session)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visualGraph)
}

// getSessionUploadProgress returns all upload progress for a session
func (h *Handlers) getSessionUploadProgress(sessionID string) map[string]*UploadProgress {
	progress := make(map[string]*UploadProgress)
	h.mutex.RLock()
	for _, uploadProgress := range h.uploadProgress {
		if uploadProgress.SessionID == sessionID {
			progress[uploadProgress.FieldID] = uploadProgress
		}
	}
	h.mutex.RUnlock()
	return progress
}

// convertToVisualGraph converts graph to visual format for admin dashboard
func (h *Handlers) convertToVisualGraph(graph *onboarding.Graph) map[string]interface{} {
	nodes := make([]map[string]interface{}, 0, len(graph.Nodes))
	edges := make([]map[string]interface{}, 0, len(graph.Edges))

	// Convert nodes
	for _, node := range graph.Nodes {
		nodeData := map[string]interface{}{
			"id":          node.ID,
			"name":        node.Name,
			"description": node.Description,
			"type":        node.Type,
			"fields":      node.Fields,
			"validation":  node.Validation,
		}
		nodes = append(nodes, nodeData)
	}

	// Convert edges
	for _, edge := range graph.Edges {
		edgeData := map[string]interface{}{
			"id":          edge.ID,
			"from":        edge.FromNodeID,
			"to":          edge.ToNodeID,
			"condition":   edge.Condition,
			"description": edge.Condition.Type, // Use condition type as description
		}
		edges = append(edges, edgeData)
	}

	return map[string]interface{}{
		"id":          graph.ID,
		"name":        graph.Name,
		"description": graph.Description,
		"nodes":       nodes,
		"edges":       edges,
		"created_at":  graph.CreatedAt,
		"updated_at":  graph.UpdatedAt,
	}
}

// convertToSessionVisualGraph converts graph to visual format with session path highlighting
func (h *Handlers) convertToSessionVisualGraph(graph *onboarding.Graph, session *onboarding.Session) map[string]interface{} {
	nodes := make([]map[string]interface{}, 0, len(graph.Nodes))
	edges := make([]map[string]interface{}, 0, len(graph.Edges))

	// Track visited nodes and edges from session history
	visitedNodes := make(map[string]bool)
	visitedEdges := make(map[string]bool)

	// Mark current node as visited
	if session.CurrentNodeID != "" {
		visitedNodes[session.CurrentNodeID] = true
	}

	// Mark nodes as visited based on session data
	for fieldName := range session.Data {
		// Find which node this field belongs to
		for nodeID, node := range graph.Nodes {
			for _, field := range node.Fields {
				if field.ID == fieldName {
					visitedNodes[nodeID] = true
					break
				}
			}
		}
	}

	// Mark edges as visited based on session history
	// Since we don't have explicit edge tracking in history, we'll infer visited edges
	// by looking at consecutive forward actions in the history
	for i := 0; i < len(session.History)-1; i++ {
		currentStep := session.History[i]
		nextStep := session.History[i+1]

		if currentStep.Action == "forward" && nextStep.Action == "forward" {
			// Find the edge ID for this transition
			for edgeID, edge := range graph.Edges {
				if edge.FromNodeID == currentStep.NodeID && edge.ToNodeID == nextStep.NodeID {
					visitedEdges[edgeID] = true
					break
				}
			}
		}
	}

	// Convert nodes with session-specific highlighting
	for _, node := range graph.Nodes {
		nodeData := map[string]interface{}{
			"id":          node.ID,
			"name":        node.Name,
			"description": node.Description,
			"type":        node.Type,
			"fields":      node.Fields,
			"validation":  node.Validation,
			"visited":     visitedNodes[node.ID],
			"current":     node.ID == session.CurrentNodeID,
		}
		nodes = append(nodes, nodeData)
	}

	// Convert edges with session-specific highlighting
	for _, edge := range graph.Edges {
		edgeData := map[string]interface{}{
			"id":          edge.ID,
			"from":        edge.FromNodeID,
			"to":          edge.ToNodeID,
			"condition":   edge.Condition,
			"description": edge.Condition.Type,
			"visited":     visitedEdges[edge.ID],
		}
		edges = append(edges, edgeData)
	}

	return map[string]interface{}{
		"id":             graph.ID,
		"name":           graph.Name,
		"description":    graph.Description,
		"nodes":          nodes,
		"edges":          edges,
		"session_id":     session.ID,
		"session_status": session.Status,
		"current_node":   session.CurrentNodeID,
		"created_at":     graph.CreatedAt,
		"updated_at":     graph.UpdatedAt,
	}
}

// buildComprehensiveNodeData builds comprehensive node data with historical information
func (h *Handlers) buildComprehensiveNodeData(session *onboarding.Session, history []onboarding.SessionStep, graph *onboarding.Graph) map[string]interface{} {
	nodeData := make(map[string]interface{})

	// Process each node in the graph
	for nodeID, node := range graph.Nodes {
		nodeInfo := map[string]interface{}{
			"id":          node.ID,
			"name":        node.Name,
			"type":        node.Type,
			"description": node.Description,
			"fields":      node.Fields,
			"visited":     false,
			"data":        make(map[string]interface{}),
			"visit_time":  nil,
		}

		// Check if this node was visited in history
		for _, step := range history {
			if step.NodeID == nodeID && step.Action == "forward" {
				nodeInfo["visited"] = true
				nodeInfo["visit_time"] = step.Timestamp

				// Extract data for this node from session data
				if step.Data != nil {
					nodeInfo["data"] = step.Data
				}
				break
			}
		}

		// Check if this node has uploaded files (consider it visited if files are uploaded)
		uploadedFiles := h.getSessionUploadedFiles(session.ID)
		for _, file := range uploadedFiles {
			// Check if any field in this node has uploaded files
			for _, field := range node.Fields {
				if field.ID == file["field_id"] {
					nodeInfo["visited"] = true
					if nodeInfo["visit_time"] == nil {
						nodeInfo["visit_time"] = file["uploaded_at"]
					}
					break
				}
			}
		}

		// If current node, also include current session data
		if nodeID == session.CurrentNodeID {
			nodeInfo["is_current"] = true
			if session.Data != nil {
				// Check if current session data contains fields that belong to this node
				currentData := session.Data
				nodeHasData := false

				// Check if any field in current data belongs to this node
				for fieldID := range currentData {
					for _, field := range node.Fields {
						if field.ID == fieldID {
							nodeHasData = true
							break
						}
					}
					if nodeHasData {
						break
					}
				}

				// Only merge data if it belongs to this node
				if nodeHasData {
					if historicalData, ok := nodeInfo["data"].(map[string]interface{}); ok {
						// Merge current data into historical data
						for k, v := range currentData {
							historicalData[k] = v
						}
						nodeInfo["data"] = historicalData
					} else {
						nodeInfo["data"] = currentData
					}

					// Mark as visited if current node has data
					if !nodeInfo["visited"].(bool) {
						nodeInfo["visited"] = true
						if nodeInfo["visit_time"] == nil {
							nodeInfo["visit_time"] = session.UpdatedAt
						}
					}
				}
			}
		}

		nodeData[nodeID] = nodeInfo
	}

	return nodeData
}

// getSessionUploadedFiles returns information about uploaded files for a session
func (h *Handlers) getSessionUploadedFiles(sessionID string) []map[string]interface{} {
	var uploadedFiles []map[string]interface{}

	// Check uploads directory for this session
	sessionUploadDir := filepath.Join(h.uploadDir, sessionID)
	if _, err := os.Stat(sessionUploadDir); os.IsNotExist(err) {
		return uploadedFiles
	}

	// Walk through the session upload directory
	err := filepath.Walk(sessionUploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Extract field ID from path
			relPath, _ := filepath.Rel(sessionUploadDir, path)
			pathParts := strings.Split(relPath, string(filepath.Separator))
			if len(pathParts) >= 2 {
				fieldID := pathParts[0]
				fileName := pathParts[1]

				fileInfo := map[string]interface{}{
					"field_id":    fieldID,
					"file_name":   fileName,
					"file_path":   path,
					"file_size":   info.Size(),
					"uploaded_at": info.ModTime(),
				}

				uploadedFiles = append(uploadedFiles, fileInfo)
			}
		}

		return nil
	})

	if err != nil {
		h.logger.WithError(err).Error("Failed to walk upload directory")
	}

	return uploadedFiles
}

// calculateSessionProgress calculates the progress percentage for a session
func (h *Handlers) calculateSessionProgress(session *onboarding.Session) int {
	// Get the graph to count total nodes
	graph, err := h.onboardingService.GetGraph(context.Background(), session.GraphID)
	if err != nil {
		return 0
	}

	totalNodes := len(graph.Nodes)
	if totalNodes == 0 {
		return 0
	}

	// Count completed nodes based on history
	completedNodes := make(map[string]bool)
	for _, step := range session.History {
		if step.Action == "forward" { // Consider forward action as completed
			completedNodes[step.NodeID] = true
		}
	}

	// Also count current node if it has data (user has filled it but not moved forward yet)
	if len(session.Data) > 0 {
		completedNodes[session.CurrentNodeID] = true
	}

	// If session is completed, return 100%
	if session.Status == "completed" {
		return 100
	}

	// Calculate progress based on completed nodes
	progress := (len(completedNodes) * 100) / totalNodes
	if progress > 100 {
		progress = 100
	}

	return progress
}

// GetEligibleNodes returns the list of nodes that are eligible for navigation based on current session state
func (h *Handlers) GetEligibleNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	// Get eligible nodes from service
	eligibleNodes, err := h.onboardingService.GetEligibleNodes(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get eligible nodes")
		http.Error(w, "Failed to get eligible nodes", http.StatusInternalServerError)
		return
	}

	// Return eligible node IDs
	response := map[string]interface{}{
		"eligible_nodes": eligibleNodes,
		"session_id":     sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadFile handles file downloads for admin
func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	// Security check: ensure the file is within the uploads directory
	fullPath := filepath.Join(h.uploadDir, filePath)
	if !strings.HasPrefix(fullPath, h.uploadDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Get file info
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get file info")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set headers for file download
	fileName := filepath.Base(fullPath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Open and serve the file
	file, err := os.Open(fullPath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to open file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy file to response
	_, err = io.Copy(w, file)
	if err != nil {
		h.logger.WithError(err).Error("Failed to copy file to response")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"file_path": fullPath,
		"file_name": fileName,
		"file_size": fileInfo.Size(),
	}).Info("File downloaded successfully")
}

// ServeIndex serves the main index page
func (h *Handlers) ServeIndex(w http.ResponseWriter, r *http.Request) {
	// Redirect to the main UI
	http.Redirect(w, r, "/dynamic-onboarding-ui.html", http.StatusFound)
}

// ServeHTML serves HTML files
func (h *Handlers) ServeHTML(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set content type
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Read and serve the HTML file
		content, err := os.ReadFile(filename)
		if err != nil {
			h.logger.WithError(err).WithField("filename", filename).Error("Failed to read HTML file")
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		w.Write(content)
	}
}
