package examples

import (
	"fmt"
	"strings"

	"onboarding-system/internal/onboarding"

	"github.com/sirupsen/logrus"
)

// DemoDynamicOnboarding demonstrates the new dynamic onboarding system
func DemoDynamicOnboarding() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create dynamic engine
	dynamicEngine := onboarding.NewDynamicEngine(logger)

	// Create production onboarding graph
	graph := CreateProductionOnboardingGraph()

	// Convert to dynamic graph for "individual" business type
	dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, "individual")

	fmt.Println("=== Dynamic Onboarding System Demo ===")
	fmt.Printf("Graph: %s\n", graph.Name)
	fmt.Printf("Total nodes: %d\n", len(dynamicGraph.DynamicNodes))
	fmt.Println()

	// Show initial node statuses
	fmt.Println("Initial Node Statuses:")
	fmt.Println("=====================")
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		fmt.Printf("- %s: %s (initial: %s)\n",
			dynamicNode.Name,
			dynamicNode.Status,
			dynamicNode.InitialStatus)

		if len(dynamicNode.Dependencies) > 0 {
			fmt.Printf("  Dependencies: %d\n", len(dynamicNode.Dependencies))
			for _, dep := range dynamicNode.Dependencies {
				fmt.Printf("    - %s %s %v (%s)\n",
					dep.FieldID, dep.Operator, dep.Value, dep.Condition)
			}
		}
	}
	fmt.Println()

	// Simulate user actions
	fmt.Println("Simulating User Actions:")
	fmt.Println("=======================")

	// 1. User selects business type
	fmt.Println("1. User selects business type: 'individual'")
	sessionData := map[string]interface{}{
		"business_type": "individual",
	}

	// Update business type and recalculate statuses
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Name == "Business Type Selection" {
			dynamicGraph.OnNodeCompleted(dynamicNode.ID, sessionData)
		}
	}
	dynamicGraph.OnNodeDataChanged("business_type_selection", "business_type", "individual", sessionData)

	fmt.Println("   Node statuses after business type selection:")
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Status != dynamicNode.InitialStatus {
			fmt.Printf("   - %s: %s -> %s\n",
				dynamicNode.Name,
				dynamicNode.InitialStatus,
				dynamicNode.Status)
		}
	}
	fmt.Println()

	// 2. User fills PAN number
	fmt.Println("2. User fills PAN number")
	sessionData["pan_number"] = "ABCDE1234F"
	sessionData["pan_document"] = "pan_doc.pdf"

	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Name == "PAN Number" {
			dynamicGraph.OnNodeCompleted(dynamicNode.ID, sessionData)
		}
	}
	dynamicGraph.OnNodeDataChanged("pan_number", "pan_number", "ABCDE1234F", sessionData)

	fmt.Println("   Node statuses after PAN completion:")
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Status == "completed" {
			fmt.Printf("   - %s: %s\n", dynamicNode.Name, dynamicNode.Status)
		}
	}
	fmt.Println()

	// 3. User selects payment channel
	fmt.Println("3. User selects payment channel: 'website'")
	sessionData["payment_channel"] = "website"
	sessionData["website_url"] = "https://example.com"

	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Name == "Payment Channel" {
			dynamicGraph.OnNodeCompleted(dynamicNode.ID, sessionData)
		}
	}
	dynamicGraph.OnNodeDataChanged("payment_channel", "payment_channel", "website", sessionData)

	fmt.Println("   Node statuses after payment channel selection:")
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Status == "completed" {
			fmt.Printf("   - %s: %s\n", dynamicNode.Name, dynamicNode.Status)
		}
	}
	fmt.Println()

	// 4. Show completion status
	fmt.Println("Completion Status:")
	fmt.Println("=================")
	completionStatus := dynamicGraph.GetCompletionStatus()
	for key, value := range completionStatus {
		fmt.Printf("- %s: %v\n", key, value)
	}
	fmt.Println()

	// 5. Show which nodes are still mandatory
	fmt.Println("Remaining Mandatory Nodes:")
	fmt.Println("=========================")
	mandatoryCount := 0
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Status == "mandatory" {
			fmt.Printf("- %s\n", dynamicNode.Name)
			mandatoryCount++
		}
	}
	if mandatoryCount == 0 {
		fmt.Println("All mandatory nodes completed! User can now complete onboarding.")
	}
	fmt.Println()

	// 6. Demonstrate dependency resolution
	fmt.Println("Dependency Resolution Example:")
	fmt.Println("=============================")
	fmt.Println("If user changes payment channel to 'app':")

	// Simulate changing payment channel
	sessionData["payment_channel"] = "app"
	sessionData["android_url"] = "https://play.google.com/app"
	sessionData["ios_url"] = "https://apps.apple.com/app"

	// Notify observers of the change
	dynamicGraph.OnNodeDataChanged("payment_channel", "payment_channel", "app", sessionData)

	fmt.Println("   Node statuses after payment channel change:")
	for _, dynamicNode := range dynamicGraph.DynamicNodes {
		if dynamicNode.Name == "Payment Channel" {
			fmt.Printf("   - %s: %s (updated with app URLs)\n",
				dynamicNode.Name, dynamicNode.Status)
		}
	}
}

// DemoBusinessTypeChange demonstrates how node statuses change with business type
func DemoBusinessTypeChange() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	dynamicEngine := onboarding.NewDynamicEngine(logger)
	graph := CreateProductionOnboardingGraph()

	fmt.Println("\n=== Business Type Change Demo ===")

	businessTypes := []string{"individual", "proprietorship", "private_limited", "partnership"}

	for _, businessType := range businessTypes {
		fmt.Printf("\nBusiness Type: %s\n", businessType)
		fmt.Println(strings.Repeat("-", len(businessType)+15))

		dynamicGraph := dynamicEngine.ConvertToDynamicGraph(graph, businessType)

		mandatoryNodes := make([]string, 0)
		optionalNodes := make([]string, 0)
		dependentNodes := make([]string, 0)

		for _, dynamicNode := range dynamicGraph.DynamicNodes {
			switch dynamicNode.Status {
			case "mandatory":
				mandatoryNodes = append(mandatoryNodes, dynamicNode.Name)
			case "optional":
				optionalNodes = append(optionalNodes, dynamicNode.Name)
			case "dependent":
				dependentNodes = append(dependentNodes, dynamicNode.Name)
			}
		}

		fmt.Printf("Mandatory (%d): %v\n", len(mandatoryNodes), mandatoryNodes)
		fmt.Printf("Optional (%d): %v\n", len(optionalNodes), optionalNodes)
		fmt.Printf("Dependent (%d): %v\n", len(dependentNodes), dependentNodes)
	}
}

// RunDynamicDemo runs the complete dynamic onboarding demo
func RunDynamicDemo() {
	fmt.Println("🚀 Starting Dynamic Onboarding System Demo")
	fmt.Println("==========================================")

	DemoDynamicOnboarding()
	DemoBusinessTypeChange()

	fmt.Println("\n✅ Dynamic Onboarding Demo Complete!")
	fmt.Println("Key Benefits:")
	fmt.Println("- Nodes dynamically change status based on user actions")
	fmt.Println("- Observer pattern ensures real-time updates")
	fmt.Println("- No need for complex static rule groups")
	fmt.Println("- Flexible and maintainable architecture")
}
