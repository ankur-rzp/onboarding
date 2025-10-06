package examples

import (
	"fmt"
)

// ExampleProductionOnboarding demonstrates how to use the production onboarding graph
func ExampleProductionOnboarding() {
	// Create the production onboarding graph
	graph := CreateProductionOnboardingGraph()

	// Note: In a real application, you would initialize the service with proper storage and config
	// service := onboarding.NewService(store, config)

	fmt.Printf("Production onboarding graph created: %s\n", graph.Name)
	fmt.Printf("Graph ID: %s\n", graph.ID)
	fmt.Printf("Total nodes: %d\n", len(graph.Nodes))
	fmt.Printf("Total edges: %d\n", len(graph.Edges))

	// Display available business types
	fmt.Println("\nAvailable business types:")
	for _, node := range graph.Nodes {
		if node.Name == "Business Type Selection" {
			for _, field := range node.Fields {
				if field.ID == "business_type" {
					for i, option := range field.Options {
						fmt.Printf("%d. %s\n", i+1, option)
					}
				}
			}
			break
		}
	}

	// Show business type specific requirements
	requirements := GetBusinessTypeRequirements()
	if req, exists := requirements["private_limited"]; exists {
		fmt.Println("\nPrivate Limited specific requirements:")
		for field, requirement := range req {
			if requirement == "required" {
				fmt.Printf("- %s: %s\n", field, requirement)
			}
		}
	}
}

// ExampleBusinessTypeValidation demonstrates conditional validation
func ExampleBusinessTypeValidation() {
	requirements := GetBusinessTypeRequirements()

	fmt.Println("Business Type Requirements Summary:")
	fmt.Println("==================================")

	for businessType, reqs := range requirements {
		fmt.Printf("\n%s:\n", businessType)
		fmt.Println("  Required documents:")
		for field, requirement := range reqs {
			if requirement == "required" {
				fmt.Printf("    - %s\n", field)
			}
		}
	}

	// Example validation for different business types
	testCases := []struct {
		businessType string
		description  string
	}{
		{"individual", "Personal PAN and basic documents"},
		{"proprietorship", "Personal PAN + MSME document"},
		{"private_limited", "Business PAN + CIN + Certificate of Incorporation"},
		{"public_limited", "Business PAN + CIN + Certificate of Incorporation"},
		{"partnership", "Business PAN + Partnership Deed"},
		{"llp", "Business PAN + Certificate of Incorporation"},
		{"trust", "Business PAN + Trust Deed"},
		{"society", "Business PAN + Society Registration Certificate"},
		{"huf", "Business PAN + HUF Deed"},
	}

	fmt.Println("\nBusiness Type Descriptions:")
	fmt.Println("===========================")
	for _, tc := range testCases {
		fmt.Printf("%s: %s\n", tc.businessType, tc.description)
	}
}

// ExampleGraphStructure demonstrates the graph structure
func ExampleGraphStructure() {
	graph := CreateProductionOnboardingGraph()

	fmt.Println("Production Onboarding Graph Structure:")
	fmt.Println("=====================================")
	fmt.Printf("Graph ID: %s\n", graph.ID)
	fmt.Printf("Graph Name: %s\n", graph.Name)
	fmt.Printf("Description: %s\n", graph.Description)
	fmt.Printf("Version: %s\n", graph.Version)
	fmt.Printf("Total Nodes: %d\n", len(graph.Nodes))
	fmt.Printf("Total Edges: %d\n", len(graph.Edges))

	fmt.Println("\nNode Flow:")
	fmt.Println("==========")

	// Follow the flow from start to end
	currentNodeID := graph.StartNodeID
	step := 1

	for currentNodeID != "" {
		node := graph.Nodes[currentNodeID]
		fmt.Printf("%d. %s (%s)\n", step, node.Name, node.Type)

		// Find the next node
		nextNodeID := ""
		for _, edge := range graph.Edges {
			if edge.FromNodeID == currentNodeID {
				nextNodeID = edge.ToNodeID
				break
			}
		}

		currentNodeID = nextNodeID
		step++
	}

	fmt.Println("\nEdge Conditions:")
	fmt.Println("===============")
	for _, edge := range graph.Edges {
		fromNode := graph.Nodes[edge.FromNodeID]
		toNode := graph.Nodes[edge.ToNodeID]
		fmt.Printf("%s -> %s (condition: %s)\n",
			fromNode.Name, toNode.Name, edge.Condition.Type)
	}
}
