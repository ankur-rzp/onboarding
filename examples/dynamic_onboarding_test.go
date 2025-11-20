package examples

import (
	"testing"

	"onboarding-system/internal/onboarding"

	"github.com/sirupsen/logrus"
)

func TestDynamicOnboardingSystem(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Create dynamic engine
	dynamicEngine := onboarding.NewDynamicEngine(logger)

	// Create production onboarding graph
	graph := CreateProductionOnboardingGraph()

	// Test 1: Convert to dynamic graph
	t.Run("ConvertToDynamicGraph", func(t *testing.T) {
		dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

		if dynamicGraph == nil {
			t.Fatal("Dynamic graph should not be nil")
		}

		if len(dynamicGraph.DynamicNodes) != len(graph.Nodes) {
			t.Errorf("Expected %d dynamic nodes, got %d", len(graph.Nodes), len(dynamicGraph.DynamicNodes))
		}
	})

	// Test 2: Initial status determination
	t.Run("InitialStatusDetermination", func(t *testing.T) {
		dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

		// Check that start node is mandatory
		for _, dynamicNode := range dynamicGraph.DynamicNodes {
			if dynamicNode.Type == "start" {
				if dynamicNode.Status != "mandatory" {
					t.Errorf("Start node should be mandatory, got %s", dynamicNode.Status)
				}
			}
		}
	})

	// Test 3: Business type specific requirements
	t.Run("BusinessTypeRequirements", func(t *testing.T) {
		businessTypes := []string{"individual", "proprietorship", "private_limited"}

		for _, businessType := range businessTypes {
			dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, businessType)

			mandatoryCount := 0
			for _, dynamicNode := range dynamicGraph.DynamicNodes {
				if dynamicNode.Status == "mandatory" {
					mandatoryCount++
				}
			}

			// Each business type should have at least some mandatory nodes
			if mandatoryCount == 0 {
				t.Errorf("Business type %s should have at least one mandatory node", businessType)
			}
		}
	})

	// Test 4: Node status updates
	t.Run("NodeStatusUpdates", func(t *testing.T) {
		dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

		// Find a node to test with
		var testNodeID string
		for _, dynamicNode := range dynamicGraph.DynamicNodes {
			if dynamicNode.Name == "Business Type Selection" {
				testNodeID = dynamicNode.ID
				break
			}
		}

		if testNodeID == "" {
			t.Fatal("Could not find Business Type Selection node")
		}

		// Test status update
		dynamicGraph.UpdateNodeStatus(testNodeID, "completed", map[string]interface{}{})
		newStatus := dynamicGraph.DynamicNodes[testNodeID].Status

		if newStatus != "completed" {
			t.Errorf("Expected status to be 'completed', got %s", newStatus)
		}
	})

	// Test 5: Completion status
	t.Run("CompletionStatus", func(t *testing.T) {
		dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

		completionStatus := dynamicGraph.GetCompletionStatus()

		// Check required fields
		requiredFields := []string{"total_nodes", "mandatory_nodes", "optional_nodes", "can_complete"}
		for _, field := range requiredFields {
			if _, exists := completionStatus[field]; !exists {
				t.Errorf("Completion status should contain field: %s", field)
			}
		}

		// Check that can_complete is boolean
		if canComplete, ok := completionStatus["can_complete"].(bool); !ok {
			t.Errorf("can_complete should be boolean, got %T", completionStatus["can_complete"])
		} else if canComplete {
			t.Error("can_complete should be false initially")
		}
	})

	// Test 6: Dependency evaluation
	t.Run("DependencyEvaluation", func(t *testing.T) {
		dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

		// Test with session data
		sessionData := map[string]interface{}{
			"business_type":   "individual",
			"payment_channel": "website",
		}

		// Find a node with dependencies
		for _, dynamicNode := range dynamicGraph.DynamicNodes {
			if len(dynamicNode.Dependencies) > 0 {
				// Test dependency evaluation
				newStatus := dynamicEngine.EvaluateNodeDependencies(dynamicNode, sessionData)

				// Status should be one of the valid statuses
				validStatuses := []onboarding.NodeStatus{"mandatory", "optional", "dependent"}
				valid := false
				for _, status := range validStatuses {
					if newStatus == status {
						valid = true
						break
					}
				}

				if !valid {
					t.Errorf("Invalid status returned: %s", newStatus)
				}
				break
			}
		}
	})
}

func TestDynamicEngineIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create dynamic engine
	dynamicEngine := onboarding.NewDynamicEngine(logger)

	// Test that it extends the base engine
	if dynamicEngine.Engine == nil {
		t.Fatal("Dynamic engine should extend base engine")
	}

	// Test business type requirements
	t.Run("BusinessTypeRequirements", func(t *testing.T) {
		graph := CreateProductionOnboardingGraph()

		// Test different business types
		testCases := []struct {
			businessType      string
			expectedMandatory []string
		}{
			{
				businessType:      "individual",
				expectedMandatory: []string{"Business Type Selection", "PAN Number", "Payment Channel", "Business Information"},
			},
			{
				businessType:      "proprietorship",
				expectedMandatory: []string{"Business Type Selection", "PAN Number", "Payment Channel", "MCC & Policy Verification", "Business Information", "Bank Account Details"},
			},
		}

		for _, tc := range testCases {
			dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, tc.businessType)

			mandatoryNodes := make([]string, 0)
			for _, dynamicNode := range dynamicGraph.DynamicNodes {
				if dynamicNode.Status == "mandatory" {
					mandatoryNodes = append(mandatoryNodes, dynamicNode.Name)
				}
			}

			// Check that expected mandatory nodes are present
			for _, expected := range tc.expectedMandatory {
				found := false
				for _, actual := range mandatoryNodes {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Business type %s should have mandatory node: %s", tc.businessType, expected)
				}
			}
		}
	})
}

// Benchmark tests
func BenchmarkDynamicGraphConversion(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dynamicEngine := onboarding.NewDynamicEngine(logger)
	graph := CreateProductionOnboardingGraph()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dynamicEngine.ConvertToDynamicGraph(graph, "individual")
	}
}

func BenchmarkNodeStatusUpdate(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	dynamicEngine := onboarding.NewDynamicEngine(logger)
	graph := CreateProductionOnboardingGraph()
	dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

	// Find a node to update
	var testNodeID string
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		testNodeID = dynamicNode.ID
		break
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dynamicGraph.UpdateNodeStatus(testNodeID, "completed", map[string]interface{}{})
	}
}
