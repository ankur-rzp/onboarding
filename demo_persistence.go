package main

import (
	"context"
	"fmt"
	"time"

	"onboarding-system/examples"
	"onboarding-system/internal/config"
	"onboarding-system/internal/onboarding"
	"onboarding-system/internal/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🚀 Dynamic Onboarding Persistence Demo")
	fmt.Println("=====================================")

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create storage and service
	storage := storage.NewMemoryStorage(logger)
	config := &config.Config{}
	dynamicService := onboarding.NewDynamicService(storage, config, logger)

	// Create and save graph
	graph := examples.CreateProductionOnboardingGraph()
	if err := storage.SaveGraph(context.Background(), graph); err != nil {
		fmt.Printf("❌ Failed to save graph: %v\n", err)
		return
	}
	fmt.Printf("✅ Graph saved with ID: %s\n", graph.ID)

	// Start a dynamic session
	fmt.Println("\n📝 Starting Dynamic Session...")
	session, err := dynamicService.StartDynamicSession(context.Background(), graph.ID, "demo-user")
	if err != nil {
		fmt.Printf("❌ Failed to start session: %v\n", err)
		return
	}
	fmt.Printf("✅ Session started with ID: %s\n", session.ID)

	// Show initial state
	fmt.Println("\n🔍 Initial Dynamic State:")
	if session.DynamicState != nil {
		fmt.Printf("  Business Type: %s\n", session.DynamicState.BusinessType)
		fmt.Printf("  Total Nodes: %d\n", len(session.DynamicState.NodeStatuses))
		fmt.Printf("  Last Evaluated: %s\n", session.DynamicState.LastEvaluatedAt.Format(time.RFC3339))
	}

	// Submit business type selection
	fmt.Println("\n📤 Submitting Business Type Selection...")
	data := map[string]interface{}{
		"business_type": "proprietorship",
	}

	result, err := dynamicService.SubmitNodeDataDynamic(context.Background(), session.ID, data)
	if err != nil {
		fmt.Printf("❌ Failed to submit data: %v\n", err)
		return
	}
	fmt.Printf("✅ Data submitted successfully\n")
	fmt.Printf("  Next Node: %s\n", result.NextNodeID)

	// Get updated session to show persistence
	updatedSession, err := storage.GetSession(context.Background(), session.ID)
	if err != nil {
		fmt.Printf("❌ Failed to get updated session: %v\n", err)
		return
	}

	fmt.Println("\n💾 Persisted Dynamic State:")
	if updatedSession.DynamicState != nil {
		fmt.Printf("  Business Type: %s\n", updatedSession.DynamicState.BusinessType)
		fmt.Printf("  Total Nodes: %d\n", len(updatedSession.DynamicState.NodeStatuses))
		fmt.Printf("  Last Evaluated: %s\n", updatedSession.DynamicState.LastEvaluatedAt.Format(time.RFC3339))

		// Show some node statuses
		fmt.Println("  Node Statuses:")
		count := 0
		for nodeID, nodeStatus := range updatedSession.DynamicState.NodeStatuses {
			if count < 5 { // Show first 5 nodes
				fmt.Printf("    - %s: %s (initial: %s)\n", nodeID, nodeStatus.Status, nodeStatus.InitialStatus)
				count++
			}
		}
		if len(updatedSession.DynamicState.NodeStatuses) > 5 {
			fmt.Printf("    ... and %d more nodes\n", len(updatedSession.DynamicState.NodeStatuses)-5)
		}
	}

	// Submit more data
	fmt.Println("\n📤 Submitting PAN Number...")
	panData := map[string]interface{}{
		"pan_number":   "ABCDE1234F",
		"pan_document": "pan.pdf",
	}

	_, err = dynamicService.SubmitNodeDataDynamic(context.Background(), session.ID, panData)
	if err != nil {
		fmt.Printf("❌ Failed to submit PAN data: %v\n", err)
		return
	}
	fmt.Printf("✅ PAN data submitted successfully\n")

	// Get dynamic node status
	fmt.Println("\n📊 Current Dynamic Node Status:")
	status, err := dynamicService.GetDynamicNodeStatus(context.Background(), session.ID)
	if err != nil {
		fmt.Printf("❌ Failed to get node status: %v\n", err)
		return
	}

	fmt.Printf("  Business Type: %s\n", status["business_type"])
	fmt.Printf("  Total Nodes: %d\n", len(status["nodes"].(map[string]interface{})))

	// Show completion status
	if completionStatus, ok := status["completion_status"].(map[string]interface{}); ok {
		fmt.Printf("  Can Complete: %v\n", completionStatus["can_complete"])
		fmt.Printf("  Completed Nodes: %v\n", completionStatus["completed_nodes"])
		fmt.Printf("  Mandatory Nodes: %v\n", completionStatus["mandatory_nodes"])
	}

	// Simulate session reload
	fmt.Println("\n🔄 Simulating Session Reload...")

	// Create a new service instance (simulating server restart)
	newDynamicService := onboarding.NewDynamicService(storage, config, logger)

	// Start a new session with the same graph (this will restore the existing session)
	reloadedSession, err := newDynamicService.StartDynamicSession(context.Background(), graph.ID, "demo-user")
	if err != nil {
		fmt.Printf("❌ Failed to reload session: %v\n", err)
		return
	}

	fmt.Printf("✅ Session reloaded with ID: %s\n", reloadedSession.ID)

	// Verify state is restored
	fmt.Println("\n🔍 Restored Dynamic State:")
	if reloadedSession.DynamicState != nil {
		fmt.Printf("  Business Type: %s\n", reloadedSession.DynamicState.BusinessType)
		fmt.Printf("  Total Nodes: %d\n", len(reloadedSession.DynamicState.NodeStatuses))
		fmt.Printf("  Last Evaluated: %s\n", reloadedSession.DynamicState.LastEvaluatedAt.Format(time.RFC3339))

		// Verify data is still there
		if reloadedSession.Data["business_type"] == "proprietorship" {
			fmt.Println("  ✅ Business type data restored")
		}
		if reloadedSession.Data["pan_number"] == "ABCDE1234F" {
			fmt.Println("  ✅ PAN number data restored")
		}
	}

	// Test business type change
	fmt.Println("\n🔄 Testing Business Type Change...")
	err = newDynamicService.UpdateBusinessTypeDynamic(context.Background(), reloadedSession.ID, "private_limited")
	if err != nil {
		fmt.Printf("❌ Failed to update business type: %v\n", err)
		return
	}
	fmt.Printf("✅ Business type updated to: private_limited\n")

	// Get final state summary
	fmt.Println("\n📋 Final State Summary:")
	summary, err := newDynamicService.GetDynamicStateSummary(context.Background(), reloadedSession.ID)
	if err != nil {
		fmt.Printf("❌ Failed to get state summary: %v\n", err)
		return
	}

	fmt.Printf("  Has Dynamic State: %v\n", summary["has_dynamic_state"])
	fmt.Printf("  Business Type: %s\n", summary["business_type"])
	fmt.Printf("  Total Nodes: %v\n", summary["total_nodes"])

	if statusCounts, ok := summary["status_counts"].(map[string]int); ok {
		fmt.Println("  Status Counts:")
		for status, count := range statusCounts {
			fmt.Printf("    - %s: %d\n", status, count)
		}
	}

	if issues, ok := summary["validation_issues"].([]string); ok && len(issues) > 0 {
		fmt.Println("  Validation Issues:")
		for _, issue := range issues {
			fmt.Printf("    - %s\n", issue)
		}
	} else {
		fmt.Println("  ✅ No validation issues found")
	}

	fmt.Println("\n🎉 Persistence Demo Complete!")
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("- ✅ Dynamic state persistence in session")
	fmt.Println("- ✅ Node status restoration on reload")
	fmt.Println("- ✅ Business type change with recalculation")
	fmt.Println("- ✅ State validation and summary")
	fmt.Println("- ✅ Session data consistency")
}
