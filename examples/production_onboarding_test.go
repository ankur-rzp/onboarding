package examples

import (
	"testing"

	"onboarding-system/internal/types"
)

func TestCreateProductionOnboardingGraph(t *testing.T) {
	graph := CreateProductionOnboardingGraph()

	// Test basic graph structure
	if graph == nil {
		t.Fatal("Graph should not be nil")
	}

	if graph.Name != "Production Onboarding" {
		t.Errorf("Expected graph name 'Production Onboarding', got '%s'", graph.Name)
	}

	if graph.StartNodeID == "" {
		t.Fatal("Start node ID should not be empty")
	}

	// Test that all expected nodes exist
	expectedNodes := []string{
		"Business Type Selection",
		"PAN Number",
		"Payment Channel",
		"MCC & Policy Verification",
		"Business Document",
		"BMC Document",
		"Authorised Signatory Details",
		"Bank Account Details",
		"Business Information",
		"Onboarding Complete",
	}

	for _, expectedName := range expectedNodes {
		found := false
		for _, node := range graph.Nodes {
			if node.Name == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected node '%s' not found in graph", expectedName)
		}
	}

	// Test that we have the expected number of nodes
	if len(graph.Nodes) != len(expectedNodes) {
		t.Errorf("Expected %d nodes, got %d", len(expectedNodes), len(graph.Nodes))
	}

	// Test that we have the expected number of edges (should be len(expectedNodes) - 1)
	expectedEdges := len(expectedNodes) - 1
	if len(graph.Edges) != expectedEdges {
		t.Errorf("Expected %d edges, got %d", expectedEdges, len(graph.Edges))
	}

	// Test business type selection node
	var startNode *types.Node
	for _, node := range graph.Nodes {
		if node.Name == "Business Type Selection" {
			startNode = node
			break
		}
	}

	if startNode == nil {
		t.Fatal("Start node not found")
	}

	if startNode.Type != types.NodeTypeStart {
		t.Errorf("Expected start node type, got %s", startNode.Type)
	}

	// Test that business type field has all expected options
	var businessTypeField *types.Field
	for _, field := range startNode.Fields {
		if field.ID == "business_type" {
			businessTypeField = &field
			break
		}
	}

	if businessTypeField == nil {
		t.Fatal("Business type field not found")
	}

	expectedBusinessTypes := []string{
		"individual", "proprietorship", "private_limited", "public_limited",
		"partnership", "llp", "trust", "society", "huf",
	}

	if len(businessTypeField.Options) != len(expectedBusinessTypes) {
		t.Errorf("Expected %d business type options, got %d", len(expectedBusinessTypes), len(businessTypeField.Options))
	}

	for _, expectedType := range expectedBusinessTypes {
		found := false
		for _, option := range businessTypeField.Options {
			if option == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected business type '%s' not found in options", expectedType)
		}
	}
}

func TestGetBusinessTypeRequirements(t *testing.T) {
	requirements := GetBusinessTypeRequirements()

	// Test that all business types are covered
	expectedBusinessTypes := []string{
		"individual", "proprietorship", "private_limited", "public_limited",
		"partnership", "llp", "trust", "society", "huf",
	}

	for _, businessType := range expectedBusinessTypes {
		if _, exists := requirements[businessType]; !exists {
			t.Errorf("Requirements for business type '%s' not found", businessType)
		}
	}

	// Test specific requirements
	testCases := []struct {
		businessType string
		field        string
		expected     string
	}{
		{"proprietorship", "msme_document", "required"},
		{"private_limited", "cin_document", "required"},
		{"public_limited", "certificate_of_incorporation", "required"},
		{"partnership", "partnership_deed", "required"},
		{"llp", "certificate_of_incorporation", "required"},
		{"trust", "trust_deed", "required"},
		{"society", "society_registration_certificate", "required"},
		{"huf", "huf_deed", "required"},
		{"individual", "msme_document", "optional"},
		{"individual", "cin_document", "optional"},
	}

	for _, tc := range testCases {
		if req, exists := requirements[tc.businessType]; exists {
			if fieldReq, fieldExists := req[tc.field]; fieldExists {
				if fieldReq != tc.expected {
					t.Errorf("For business type '%s', field '%s' expected '%s', got '%s'",
						tc.businessType, tc.field, tc.expected, fieldReq)
				}
			} else {
				t.Errorf("Field '%s' not found in requirements for business type '%s'", tc.field, tc.businessType)
			}
		} else {
			t.Errorf("Business type '%s' not found in requirements", tc.businessType)
		}
	}
}

func TestGraphFlow(t *testing.T) {
	graph := CreateProductionOnboardingGraph()

	// Test that the flow is linear (each node has exactly one outgoing edge except the last)
	nodeOutgoingEdges := make(map[string]int)
	nodeIncomingEdges := make(map[string]int)

	for _, edge := range graph.Edges {
		nodeOutgoingEdges[edge.FromNodeID]++
		nodeIncomingEdges[edge.ToNodeID]++
	}

	// Find the completion node
	var completionNode *types.Node
	for _, node := range graph.Nodes {
		if node.Name == "Onboarding Complete" {
			completionNode = node
			break
		}
	}

	if completionNode == nil {
		t.Fatal("Completion node not found")
	}

	// Test that completion node has no outgoing edges
	if nodeOutgoingEdges[completionNode.ID] != 0 {
		t.Error("Completion node should have no outgoing edges")
	}

	// Test that start node has no incoming edges
	var startNode *types.Node
	for _, node := range graph.Nodes {
		if node.Name == "Business Type Selection" {
			startNode = node
			break
		}
	}

	if startNode == nil {
		t.Fatal("Start node not found")
	}

	if nodeIncomingEdges[startNode.ID] != 0 {
		t.Error("Start node should have no incoming edges")
	}

	// Test that all other nodes have exactly one incoming and one outgoing edge
	for _, node := range graph.Nodes {
		if node.ID == startNode.ID || node.ID == completionNode.ID {
			continue
		}

		if nodeIncomingEdges[node.ID] != 1 {
			t.Errorf("Node '%s' should have exactly 1 incoming edge, got %d", node.Name, nodeIncomingEdges[node.ID])
		}

		if nodeOutgoingEdges[node.ID] != 1 {
			t.Errorf("Node '%s' should have exactly 1 outgoing edge, got %d", node.Name, nodeOutgoingEdges[node.ID])
		}
	}
}

func TestValidationRules(t *testing.T) {
	graph := CreateProductionOnboardingGraph()

	// Test that business document node has conditional validation rules
	var businessDocumentNode *types.Node
	for _, node := range graph.Nodes {
		if node.Name == "Business Document" {
			businessDocumentNode = node
			break
		}
	}

	if businessDocumentNode == nil {
		t.Fatal("Business document node not found")
	}

	if len(businessDocumentNode.Validation.Conditions) == 0 {
		t.Error("Business document node should have conditional validation rules")
	}

	// Test that PAN node has required fields
	var panNode *types.Node
	for _, node := range graph.Nodes {
		if node.Name == "PAN Number" {
			panNode = node
			break
		}
	}

	if panNode == nil {
		t.Fatal("PAN node not found")
	}

	if len(panNode.Validation.RequiredFields) == 0 {
		t.Error("PAN node should have required fields")
	}

	// Test that PAN number field has proper validation
	var panNumberField *types.Field
	for _, field := range panNode.Fields {
		if field.ID == "pan_number" {
			panNumberField = &field
			break
		}
	}

	if panNumberField == nil {
		t.Fatal("PAN number field not found")
	}

	if panNumberField.Validation.Pattern == "" {
		t.Error("PAN number field should have validation pattern")
	}
}

