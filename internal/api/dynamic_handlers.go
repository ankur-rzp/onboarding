package api

import (
	"encoding/json"
	"net/http"

	"onboarding-system/internal/onboarding"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DynamicHandlers handles dynamic onboarding API endpoints
type DynamicHandlers struct {
	dynamicService *onboarding.DynamicService
	logger         *logrus.Logger
}

// NewDynamicHandlers creates a new dynamic handlers instance
func NewDynamicHandlers(dynamicService *onboarding.DynamicService, logger *logrus.Logger) *DynamicHandlers {
	return &DynamicHandlers{
		dynamicService: dynamicService,
		logger:         logger,
	}
}

// StartDynamicSession starts a new dynamic session
func (dh *DynamicHandlers) StartDynamicSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GraphID string `json:"graph_id"`
		UserID  string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dh.logger.WithError(err).Error("Failed to decode start dynamic session request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.GraphID == "" || req.UserID == "" {
		http.Error(w, "graph_id and user_id are required", http.StatusBadRequest)
		return
	}

	session, err := dh.dynamicService.StartDynamicSession(r.Context(), req.GraphID, req.UserID)
	if err != nil {
		dh.logger.WithError(err).Error("Failed to start dynamic session")
		http.Error(w, "Failed to start session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// SubmitNodeDataDynamic submits node data with dynamic status updates
func (dh *DynamicHandlers) SubmitNodeDataDynamic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		dh.logger.WithError(err).Error("Failed to decode submit node data request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := dh.dynamicService.SubmitNodeDataDynamic(r.Context(), sessionID, data)
	if err != nil {
		dh.logger.WithError(err).Error("Failed to submit node data")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetDynamicNodeStatus returns the current status of all nodes in a session
func (dh *DynamicHandlers) GetDynamicNodeStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	status, err := dh.dynamicService.GetDynamicNodeStatus(r.Context(), sessionID)
	if err != nil {
		dh.logger.WithError(err).Error("Failed to get dynamic node status")
		http.Error(w, "Failed to get node status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// UpdateBusinessType updates the business type and recalculates node statuses
func (dh *DynamicHandlers) UpdateBusinessType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	var req struct {
		BusinessType string `json:"business_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dh.logger.WithError(err).Error("Failed to decode update business type request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.BusinessType == "" {
		http.Error(w, "business_type is required", http.StatusBadRequest)
		return
	}

	err := dh.dynamicService.UpdateBusinessTypeDynamic(r.Context(), sessionID, req.BusinessType)
	if err != nil {
		dh.logger.WithError(err).Error("Failed to update business type")
		http.Error(w, "Failed to update business type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetEligibleNodesDynamic returns nodes that are eligible for navigation based on dynamic status
func (dh *DynamicHandlers) GetEligibleNodesDynamic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	eligibleNodes, err := dh.dynamicService.GetEligibleNodesDynamic(r.Context(), sessionID)
	if err != nil {
		dh.logger.WithError(err).Error("Failed to get eligible nodes")
		http.Error(w, "Failed to get eligible nodes", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"eligible_nodes": eligibleNodes,
		"session_id":     sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterDynamicRoutes registers dynamic API routes
func (dh *DynamicHandlers) RegisterDynamicRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/dynamic").Subrouter()

	// Test endpoint
	api.HandleFunc("/test", dh.testEndpoint).Methods("GET")

	// Dynamic session management
	api.HandleFunc("/sessions", dh.StartDynamicSession).Methods("POST")
	api.HandleFunc("/sessions/{id}/submit", dh.SubmitNodeDataDynamic).Methods("POST")
	api.HandleFunc("/sessions/{id}/status", dh.GetDynamicNodeStatus).Methods("GET")
	api.HandleFunc("/sessions/{id}/business-type", dh.UpdateBusinessTypeDynamic).Methods("PUT")
	api.HandleFunc("/sessions/{id}/eligible-nodes", dh.GetEligibleNodesDynamic).Methods("GET")
	api.HandleFunc("/sessions/{id}/summary", dh.GetDynamicStateSummary).Methods("GET")

	// CORS support
	api.HandleFunc("/test", dh.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions", dh.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/submit", dh.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/status", dh.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/business-type", dh.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/eligible-nodes", dh.corsHandler).Methods("OPTIONS")
	api.HandleFunc("/sessions/{id}/summary", dh.corsHandler).Methods("OPTIONS")
}

// testEndpoint is a simple test endpoint
func (dh *DynamicHandlers) testEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "dynamic handlers working",
		"message": "Dynamic API endpoints are registered successfully",
	})
}

// corsHandler handles CORS preflight requests
func (dh *DynamicHandlers) corsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}

// UpdateBusinessTypeDynamic updates the business type and recalculates node statuses
func (h *DynamicHandlers) UpdateBusinessTypeDynamic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["session_id"]

	var request struct {
		BusinessType string `json:"business_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.BusinessType == "" {
		http.Error(w, "Business type is required", http.StatusBadRequest)
		return
	}

	err := h.dynamicService.UpdateBusinessTypeDynamic(r.Context(), sessionID, request.BusinessType)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update business type")
		http.Error(w, "Failed to update business type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Business type updated successfully",
		"business_type": request.BusinessType,
	})
}

// GetDynamicStateSummary returns a summary of the dynamic state
func (h *DynamicHandlers) GetDynamicStateSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["session_id"]

	summary, err := h.dynamicService.GetDynamicStateSummary(r.Context(), sessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get dynamic state summary")
		http.Error(w, "Failed to get dynamic state summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
