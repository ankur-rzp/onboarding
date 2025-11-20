package examples

import (
	"context"
	"fmt"
	"testing"

	"onboarding-system/internal/config"
	"onboarding-system/internal/onboarding"
	"onboarding-system/internal/storage"

	"github.com/sirupsen/logrus"
)

func TestDynamicPersistence(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress verbose logs during tests

	// Create in-memory storage
	storage := storage.NewMemoryStorage(logger)
	config := &config.Config{}

	// Create dynamic service
	dynamicService := onboarding.NewDynamicService(storage, config, logger)

	// Create production graph
	graph := CreateProductionOnboardingGraph()
	if err := storage.SaveGraph(context.Background(), graph); err != nil {
		t.Fatalf("Failed to save graph: %v", err)
	}

	t.Run("SessionPersistence", func(t *testing.T) {
		// Start a dynamic session
		session, err := dynamicService.StartDynamicSession(context.Background(), graph.ID, "test-user")
		if err != nil {
			t.Fatalf("Failed to start dynamic session: %v", err)
		}

		// Verify initial dynamic state is saved
		if session.DynamicState == nil {
			t.Fatal("Dynamic state should be initialized")
		}

		if session.DynamicState.BusinessType != "individual" {
			t.Errorf("Expected business type 'individual', got '%s'", session.DynamicState.BusinessType)
		}

		// Submit some data
		data := map[string]interface{}{
			"business_type": "individual",
			"pan_number":    "ABCDE1234F",
		}

		_, err = dynamicService.SubmitNodeDataDynamic(context.Background(), session.ID, data)
		if err != nil {
			t.Fatalf("Failed to submit node data: %v", err)
		}

		// Get updated session
		updatedSession, err := storage.GetSession(context.Background(), session.ID)
		if err != nil {
			t.Fatalf("Failed to get updated session: %v", err)
		}

		// Verify dynamic state is persisted
		if updatedSession.DynamicState == nil {
			t.Fatal("Dynamic state should be persisted")
		}

		if updatedSession.DynamicState.BusinessType != "individual" {
			t.Errorf("Expected business type 'individual', got '%s'", updatedSession.DynamicState.BusinessType)
		}

		// Verify node statuses are saved
		if len(updatedSession.DynamicState.NodeStatuses) == 0 {
			t.Fatal("Node statuses should be saved")
		}

		// Check that completion status is saved
		if updatedSession.DynamicState.CompletionStatus == nil {
			t.Fatal("Completion status should be saved")
		}
	})

	t.Run("SessionReload", func(t *testing.T) {
		// Start a session and submit data
		session, err := dynamicService.StartDynamicSession(context.Background(), graph.ID, "test-user-2")
		if err != nil {
			t.Fatalf("Failed to start dynamic session: %v", err)
		}

		// Submit business type selection
		data := map[string]interface{}{
			"business_type": "proprietorship",
		}

		_, err = dynamicService.SubmitNodeDataDynamic(context.Background(), session.ID, data)
		if err != nil {
			t.Fatalf("Failed to submit business type: %v", err)
		}

		// Simulate session reload by creating a new service instance
		newDynamicService := onboarding.NewDynamicService(storage, config, logger)

		// Get the session (simulating reload)
		_, err = storage.GetSession(context.Background(), session.ID)
		if err != nil {
			t.Fatalf("Failed to get reloaded session: %v", err)
		}

		// Start a new dynamic session with the same graph (simulating reload)
		newSession, err := newDynamicService.StartDynamicSession(context.Background(), graph.ID, "test-user-2")
		if err != nil {
			t.Fatalf("Failed to start new dynamic session: %v", err)
		}

		// The session should be the same
		if newSession.ID != session.ID {
			t.Errorf("Expected session ID %s, got %s", session.ID, newSession.ID)
		}

		// Verify dynamic state is restored
		if newSession.DynamicState == nil {
			t.Fatal("Dynamic state should be restored")
		}

		if newSession.DynamicState.BusinessType != "proprietorship" {
			t.Errorf("Expected business type 'proprietorship', got '%s'", newSession.DynamicState.BusinessType)
		}

		// Verify node statuses are restored
		if len(newSession.DynamicState.NodeStatuses) == 0 {
			t.Fatal("Node statuses should be restored")
		}
	})

	t.Run("BusinessTypeChange", func(t *testing.T) {
		// Start a session
		session, err := dynamicService.StartDynamicSession(context.Background(), graph.ID, "test-user-3")
		if err != nil {
			t.Fatalf("Failed to start dynamic session: %v", err)
		}

		// Change business type
		err = dynamicService.UpdateBusinessTypeDynamic(context.Background(), session.ID, "private_limited")
		if err != nil {
			t.Fatalf("Failed to update business type: %v", err)
		}

		// Get updated session
		updatedSession, err := storage.GetSession(context.Background(), session.ID)
		if err != nil {
			t.Fatalf("Failed to get updated session: %v", err)
		}

		// Verify business type is updated
		if updatedSession.DynamicState.BusinessType != "private_limited" {
			t.Errorf("Expected business type 'private_limited', got '%s'", updatedSession.DynamicState.BusinessType)
		}

		// Verify session data is updated
		if updatedSession.Data["business_type"] != "private_limited" {
			t.Errorf("Expected session data business_type 'private_limited', got '%v'", updatedSession.Data["business_type"])
		}
	})

	t.Run("StateValidation", func(t *testing.T) {
		// Start a session
		session, err := dynamicService.StartDynamicSession(context.Background(), graph.ID, "test-user-4")
		if err != nil {
			t.Fatalf("Failed to start dynamic session: %v", err)
		}

		// Get state summary
		summary, err := dynamicService.GetDynamicStateSummary(context.Background(), session.ID)
		if err != nil {
			t.Fatalf("Failed to get dynamic state summary: %v", err)
		}

		// Verify summary structure
		if summary["has_dynamic_state"] != true {
			t.Error("Expected has_dynamic_state to be true")
		}

		if summary["business_type"] != "individual" {
			t.Errorf("Expected business_type 'individual', got '%v'", summary["business_type"])
		}

		if summary["total_nodes"] == nil {
			t.Error("Expected total_nodes to be present")
		}

		// Check for validation issues
		if issues, ok := summary["validation_issues"].([]string); ok && len(issues) > 0 {
			t.Errorf("Unexpected validation issues: %v", issues)
		}
	})
}

func TestDynamicPersistencePerformance(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	storage := storage.NewMemoryStorage(logger)
	config := &config.Config{}
	dynamicService := onboarding.NewDynamicService(storage, config, logger)

	graph := CreateProductionOnboardingGraph()
	if err := storage.SaveGraph(context.Background(), graph); err != nil {
		t.Fatalf("Failed to save graph: %v", err)
	}

	// Test multiple sessions with persistence
	sessionCount := 10
	sessions := make([]*onboarding.Session, sessionCount)

	// Create multiple sessions
	for i := 0; i < sessionCount; i++ {
		session, err := dynamicService.StartDynamicSession(context.Background(), graph.ID, fmt.Sprintf("user-%d", i))
		if err != nil {
			t.Fatalf("Failed to start session %d: %v", i, err)
		}
		sessions[i] = session
	}

	// Submit data to all sessions
	for i, session := range sessions {
		data := map[string]interface{}{
			"business_type": "individual",
			"pan_number":    fmt.Sprintf("ABCDE%dF", i),
		}

		_, err := dynamicService.SubmitNodeDataDynamic(context.Background(), session.ID, data)
		if err != nil {
			t.Fatalf("Failed to submit data to session %d: %v", i, err)
		}
	}

	// Verify all sessions have persisted state
	for i, session := range sessions {
		updatedSession, err := storage.GetSession(context.Background(), session.ID)
		if err != nil {
			t.Fatalf("Failed to get session %d: %v", i, err)
		}

		if updatedSession.DynamicState == nil {
			t.Fatalf("Session %d should have dynamic state", i)
		}

		if len(updatedSession.DynamicState.NodeStatuses) == 0 {
			t.Fatalf("Session %d should have node statuses", i)
		}
	}
}

func BenchmarkDynamicPersistence(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	storage := storage.NewMemoryStorage(logger)
	config := &config.Config{}
	dynamicService := onboarding.NewDynamicService(storage, config, logger)

	graph := CreateProductionOnboardingGraph()
	storage.SaveGraph(context.Background(), graph)

	session, _ := dynamicService.StartDynamicSession(context.Background(), graph.ID, "bench-user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := map[string]interface{}{
			"business_type": "individual",
			"pan_number":    fmt.Sprintf("ABCDE%dF", i),
		}
		dynamicService.SubmitNodeDataDynamic(context.Background(), session.ID, data)
	}
}
