package examples

import (
	"onboarding-system/internal/types"
)

// CreateEnhancedDynamicProductionOnboardingGraph creates a dynamic onboarding graph with proper edge tracking
func CreateEnhancedDynamicProductionOnboardingGraph() *types.Graph {
	graph := types.NewGraph("Enhanced Dynamic Production Onboarding", "Dynamic onboarding flow with proper edge tracking and navigation metadata")

	// Create all nodes with proper metadata
	startNode := &types.Node{
		ID:          "business_type_selection",
		Type:        "start",
		Name:        "Business Type Selection",
		Description: "Select your business type",
		Fields: []types.Field{
			{
				ID:       "business_type",
				Name:     "business_type",
				Type:     types.FieldTypeSelect,
				Required: true,
				Options:  []string{"individual", "proprietorship", "private_limited", "public_limited", "partnership", "llp", "trust", "society", "huf"},
				Validation: types.FieldValidation{
					CustomRules: []string{"business_type_validation"},
				},
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{"business_type"},
		},
		IsIndependent: true,
		IsDependent:   false,
		Dependencies:  []types.NodeDependency{},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	panNode := &types.Node{
		ID:          "pan_number_node_id",
		Type:        "input",
		Name:        "PAN Number",
		Description: "Enter your PAN number",
		Fields: []types.Field{
			{
				ID:       "pan_number",
				Name:     "pan_number",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					Pattern:     `^[A-Z]{5}[0-9]{4}[A-Z]{1}$`,
					CustomRules: []string{"pan_validation"},
				},
			},
			{
				ID:       "pan_document",
				Name:     "pan_document",
				Type:     types.FieldTypeFile,
				Required: true,
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{"pan_number", "pan_document"},
		},
		IsIndependent: true,
		IsDependent:   false,
		Dependencies:  []types.NodeDependency{},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	paymentChannelNode := &types.Node{
		ID:          "payment_channel_node_id",
		Type:        "input",
		Name:        "Payment Channel",
		Description: "Select your payment channel",
		Fields: []types.Field{
			{
				ID:       "payment_channel",
				Name:     "payment_channel",
				Type:     types.FieldTypeSelect,
				Required: true,
				Options:  []string{"website", "app", "no_code"},
			},
			{
				ID:       "website_url",
				Name:     "website_url",
				Type:     types.FieldTypeText,
				Required: false,
				Validation: types.FieldValidation{
					Pattern: `^https?://[^\s/$.?#].[^\s]*$`,
				},
			},
			{
				ID:       "android_url",
				Name:     "android_url",
				Type:     types.FieldTypeText,
				Required: false,
				Validation: types.FieldValidation{
					Pattern: `^https?://[^\s/$.?#].[^\s]*$`,
				},
			},
			{
				ID:       "ios_url",
				Name:     "ios_url",
				Type:     types.FieldTypeText,
				Required: false,
				Validation: types.FieldValidation{
					Pattern: `^https?://[^\s/$.?#].[^\s]*$`,
				},
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{"payment_channel"},
			Conditions: []types.ValidationCondition{
				{
					Field:    "payment_channel",
					Operator: "eq",
					Value:    "website",
					Rule:     "website_url is required when website is selected",
				},
				{
					Field:    "payment_channel",
					Operator: "eq",
					Value:    "app",
					Rule:     "android_url and ios_url are required when app is selected",
				},
			},
		},
		IsIndependent: true,
		IsDependent:   false,
		Dependencies:  []types.NodeDependency{},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	mccPolicyNode := &types.Node{
		ID:          "mcc_policy_verification_node_id",
		Type:        "input",
		Name:        "MCC & Policy Verification",
		Description: "Provide business category and policy information",
		Fields: []types.Field{
			{
				ID:       "subcategory",
				Name:     "subcategory",
				Type:     types.FieldTypeSelect,
				Required: true,
				Options:  []string{"retail", "wholesale", "services", "manufacturing", "technology", "healthcare", "education", "finance", "real_estate", "hospitality", "transportation", "agriculture", "consulting", "other"},
			},
			{
				ID:       "policy_pages",
				Name:     "policy_pages",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 10,
					MaxLength: 1000,
				},
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{"subcategory", "policy_pages"},
		},
		IsIndependent: true,
		IsDependent:   false,
		Dependencies:  []types.NodeDependency{},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	businessInfoNode := &types.Node{
		ID:          "business_info_node_id",
		Type:        "input",
		Name:        "Business Information",
		Description: "Enter business details",
		Fields: []types.Field{
			{
				ID:       "business_name",
				Name:     "business_name",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 2,
					MaxLength: 200,
				},
			},
			{
				ID:       "brand_name",
				Name:     "brand_name",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 2,
					MaxLength: 200,
				},
			},
			{
				ID:       "business_address_line1",
				Name:     "business_address_line1",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 5,
					MaxLength: 200,
				},
			},
			{
				ID:       "business_address_line2",
				Name:     "business_address_line2",
				Type:     types.FieldTypeText,
				Required: false,
				Validation: types.FieldValidation{
					MaxLength: 200,
				},
			},
			{
				ID:       "business_city",
				Name:     "business_city",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 2,
					MaxLength: 50,
				},
			},
			{
				ID:       "business_state",
				Name:     "business_state",
				Type:     types.FieldTypeSelect,
				Required: true,
				Options:  []string{"andhra_pradesh", "arunachal_pradesh", "assam", "bihar", "chhattisgarh", "goa", "gujarat", "haryana", "himachal_pradesh", "jharkhand", "karnataka", "kerala", "madhya_pradesh", "maharashtra", "manipur", "meghalaya", "mizoram", "nagaland", "odisha", "punjab", "rajasthan", "sikkim", "tamil_nadu", "telangana", "tripura", "uttar_pradesh", "uttarakhand", "west_bengal"},
			},
			{
				ID:       "business_pincode",
				Name:     "business_pincode",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					Pattern: `^[1-9][0-9]{5}$`,
				},
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
		},
		IsIndependent: true,
		IsDependent:   false,
		Dependencies:  []types.NodeDependency{},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	businessDocumentNode := &types.Node{
		ID:          "business_document_node_id",
		Type:        "input",
		Name:        "Business Document",
		Description: "Upload required business documents",
		Fields: []types.Field{
			{
				ID:       "msme_document",
				Name:     "msme_document",
				Type:     types.FieldTypeFile,
				Required: false,
			},
			{
				ID:       "gst_document",
				Name:     "gst_document",
				Type:     types.FieldTypeFile,
				Required: false,
			},
			{
				ID:       "gstin_number",
				Name:     "gstin_number",
				Type:     types.FieldTypeText,
				Required: false,
				Validation: types.FieldValidation{
					Pattern:     `^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`,
					CustomRules: []string{"gst_validation"},
				},
			},
			{
				ID:       "cin_document",
				Name:     "cin_document",
				Type:     types.FieldTypeFile,
				Required: false,
			},
			{
				ID:       "cin_number",
				Name:     "cin_number",
				Type:     types.FieldTypeText,
				Required: false,
				Validation: types.FieldValidation{
					Pattern: `^[A-Z]{2}[0-9]{2}[A-Z]{2}[0-9]{4}$`,
				},
			},
			{
				ID:       "certificate_of_incorporation",
				Name:     "certificate_of_incorporation",
				Type:     types.FieldTypeFile,
				Required: false,
			},
			{
				ID:       "partnership_deed",
				Name:     "partnership_deed",
				Type:     types.FieldTypeFile,
				Required: false,
			},
			{
				ID:       "trust_deed",
				Name:     "trust_deed",
				Type:     types.FieldTypeFile,
				Required: false,
			},
			{
				ID:       "society_registration_certificate",
				Name:     "society_registration_certificate",
				Type:     types.FieldTypeFile,
				Required: false,
			},
			{
				ID:       "huf_deed",
				Name:     "huf_deed",
				Type:     types.FieldTypeFile,
				Required: false,
			},
		},
		Validation: types.ValidationRules{
			Conditions: []types.ValidationCondition{
				{
					Field:    "business_type",
					Operator: "eq",
					Value:    "proprietorship",
					Rule:     "msme_document is required for proprietorship",
				},
				{
					Field:    "business_type",
					Operator: "in",
					Value:    "private_limited,public_limited",
					Rule:     "cin_document and certificate_of_incorporation are required for limited companies",
				},
				{
					Field:    "business_type",
					Operator: "eq",
					Value:    "partnership",
					Rule:     "partnership_deed is required for partnership",
				},
				{
					Field:    "business_type",
					Operator: "eq",
					Value:    "llp",
					Rule:     "certificate_of_incorporation is required for LLP",
				},
				{
					Field:    "business_type",
					Operator: "eq",
					Value:    "trust",
					Rule:     "trust_deed is required for trust",
				},
				{
					Field:    "business_type",
					Operator: "eq",
					Value:    "society",
					Rule:     "society_registration_certificate is required for society",
				},
				{
					Field:    "business_type",
					Operator: "eq",
					Value:    "huf",
					Rule:     "huf_deed is required for HUF",
				},
			},
		},
		IsIndependent: false,
		IsDependent:   true,
		Dependencies: []types.NodeDependency{
			{
				FieldID:      "business_type",
				Operator:     "ne",
				Value:        "",
				Condition:    "Business type must be selected",
				BusinessType: "",
			},
		},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	bmcDocumentNode := &types.Node{
		ID:          "bmc_document_node_id",
		Type:        "input",
		Name:        "BMC Document",
		Description: "Upload BMC document if required",
		Fields: []types.Field{
			{
				ID:       "bmc_document",
				Name:     "bmc_document",
				Type:     types.FieldTypeFile,
				Required: false,
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{},
		},
		IsIndependent: false,
		IsDependent:   true,
		Dependencies: []types.NodeDependency{
			{
				FieldID:      "subcategory",
				Operator:     "ne",
				Value:        "",
				Condition:    "Subcategory must be selected",
				BusinessType: "",
			},
		},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	authorisedSignatoryNode := &types.Node{
		ID:          "authorised_signatory_node_id",
		Type:        "input",
		Name:        "Authorised Signatory Details",
		Description: "Enter authorised signatory information",
		Fields: []types.Field{
			{
				ID:       "signatory_name",
				Name:     "signatory_name",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 2,
					MaxLength: 100,
				},
			},
			{
				ID:       "signatory_pan",
				Name:     "signatory_pan",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					Pattern:     `^[A-Z]{5}[0-9]{4}[A-Z]{1}$`,
					CustomRules: []string{"pan_validation"},
				},
			},
			{
				ID:       "signatory_pan_document",
				Name:     "signatory_pan_document",
				Type:     types.FieldTypeFile,
				Required: true,
			},
			{
				ID:       "signatory_aadhaar_front",
				Name:     "signatory_aadhaar_front",
				Type:     types.FieldTypeFile,
				Required: true,
			},
			{
				ID:       "signatory_aadhaar_back",
				Name:     "signatory_aadhaar_back",
				Type:     types.FieldTypeFile,
				Required: true,
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{"signatory_name", "signatory_pan", "signatory_pan_document", "signatory_aadhaar_front", "signatory_aadhaar_back"},
		},
		IsIndependent: false,
		IsDependent:   true,
		Dependencies: []types.NodeDependency{
			{
				FieldID:      "business_type",
				Operator:     "ne",
				Value:        "",
				Condition:    "Business type must be selected",
				BusinessType: "",
			},
		},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	bankAccountNode := &types.Node{
		ID:          "bank_account_node_id",
		Type:        "input",
		Name:        "Bank Account Details",
		Description: "Enter bank account information",
		Fields: []types.Field{
			{
				ID:       "bank_account_number",
				Name:     "bank_account_number",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 9,
					MaxLength: 18,
					Pattern:   `^[0-9]+$`,
				},
			},
			{
				ID:       "ifsc_code",
				Name:     "ifsc_code",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					Pattern: `^[A-Z]{4}0[A-Z0-9]{6}$`,
				},
			},
			{
				ID:       "bank_name",
				Name:     "bank_name",
				Type:     types.FieldTypeText,
				Required: true,
				Validation: types.FieldValidation{
					MinLength: 2,
					MaxLength: 100,
				},
			},
		},
		Validation: types.ValidationRules{
			RequiredFields: []string{"bank_account_number", "ifsc_code", "bank_name"},
			CustomRules:    []string{"penny_testing_verification"},
		},
		IsIndependent: false,
		IsDependent:   true,
		Dependencies: []types.NodeDependency{
			{
				FieldID:      "business_type",
				Operator:     "ne",
				Value:        "",
				Condition:    "Business type must be selected",
				BusinessType: "",
			},
		},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	completionNode := &types.Node{
		ID:            "completion_node_id",
		Type:          "end",
		Name:          "Onboarding Complete",
		Description:   "Your onboarding is complete and under review",
		Fields:        []types.Field{},
		Validation:    types.ValidationRules{},
		IsIndependent: false,
		IsDependent:   false,
		Dependencies:  []types.NodeDependency{},
		IncomingEdges: []string{},
		OutgoingEdges: []string{},
	}

	// Add all nodes to the graph
	graph.Nodes[startNode.ID] = startNode
	graph.Nodes[panNode.ID] = panNode
	graph.Nodes[paymentChannelNode.ID] = paymentChannelNode
	graph.Nodes[mccPolicyNode.ID] = mccPolicyNode
	graph.Nodes[businessInfoNode.ID] = businessInfoNode
	graph.Nodes[businessDocumentNode.ID] = businessDocumentNode
	graph.Nodes[bmcDocumentNode.ID] = bmcDocumentNode
	graph.Nodes[authorisedSignatoryNode.ID] = authorisedSignatoryNode
	graph.Nodes[bankAccountNode.ID] = bankAccountNode
	graph.Nodes[completionNode.ID] = completionNode

	// Create edges and build relationships
	edges := buildEnhancedEdges(startNode, panNode, paymentChannelNode, mccPolicyNode, businessInfoNode, businessDocumentNode, bmcDocumentNode, authorisedSignatoryNode, bankAccountNode, completionNode)

	// Add edges to graph and update node relationships
	for _, edge := range edges {
		graph.Edges[edge.ID] = edge

		// Update node edge relationships
		fromNode := graph.Nodes[edge.FromNodeID]
		toNode := graph.Nodes[edge.ToNodeID]

		if fromNode != nil {
			fromNode.OutgoingEdges = append(fromNode.OutgoingEdges, edge.ID)
		}
		if toNode != nil {
			toNode.IncomingEdges = append(toNode.IncomingEdges, edge.ID)
		}
	}

	// Set the start node
	graph.StartNodeID = startNode.ID

	// Add cross-node validation rules
	graph.CrossNodeValidation = buildCrossNodeValidationRules()

	return graph
}

// buildEnhancedEdges creates all edges with proper relationships
func buildEnhancedEdges(startNode, panNode, paymentChannelNode, mccPolicyNode, businessInfoNode, businessDocumentNode, bmcDocumentNode, authorisedSignatoryNode, bankAccountNode, completionNode *types.Node) []*types.Edge {
	var edges []*types.Edge

	// Start node to independent nodes
	edges = append(edges, types.NewEdge(startNode.ID, panNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(startNode.ID, paymentChannelNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(startNode.ID, mccPolicyNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(startNode.ID, businessInfoNode.ID, types.EdgeCondition{Type: "always"}))

	// Independent nodes to each other
	edges = append(edges, types.NewEdge(panNode.ID, paymentChannelNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(panNode.ID, mccPolicyNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(panNode.ID, businessInfoNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(paymentChannelNode.ID, mccPolicyNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(paymentChannelNode.ID, businessInfoNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(mccPolicyNode.ID, businessInfoNode.ID, types.EdgeCondition{Type: "always"}))

	// Independent nodes to dependent nodes (with conditions)
	edges = append(edges, types.NewEdge(startNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(panNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(paymentChannelNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(mccPolicyNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(businessInfoNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))

	// MCC Policy to BMC Document (depends on subcategory)
	edges = append(edges, types.NewEdge(mccPolicyNode.ID, bmcDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "subcategory",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(businessDocumentNode.ID, bmcDocumentNode.ID, types.EdgeCondition{Type: "always"}))

	// Independent nodes to Authorised Signatory (depends on business_type)
	edges = append(edges, types.NewEdge(startNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(panNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(paymentChannelNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(mccPolicyNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(businessInfoNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(businessDocumentNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{Type: "always"}))
	edges = append(edges, types.NewEdge(bmcDocumentNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{Type: "always"}))

	// Independent nodes to Bank Account (depends on business_type)
	edges = append(edges, types.NewEdge(startNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(panNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(paymentChannelNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(mccPolicyNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(businessInfoNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	}))
	edges = append(edges, types.NewEdge(authorisedSignatoryNode.ID, bankAccountNode.ID, types.EdgeCondition{Type: "always"}))

	// All nodes to completion (with completion check)
	edges = append(edges, types.NewEdge(startNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(panNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(paymentChannelNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(mccPolicyNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(businessInfoNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(businessDocumentNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(bmcDocumentNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(authorisedSignatoryNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))
	edges = append(edges, types.NewEdge(bankAccountNode.ID, completionNode.ID, types.EdgeCondition{Type: "completion_check"}))

	return edges
}

// buildCrossNodeValidationRules creates cross-node validation rules
func buildCrossNodeValidationRules() []types.CrossNodeValidationRule {
	return []types.CrossNodeValidationRule{
		{
			ID:          "business_name_bank_name_match",
			Name:        "Business Name and Bank Name Consistency",
			Description: "Ensures that the business name matches or is consistent with the bank name",
			Fields: []types.CrossNodeFieldReference{
				{
					NodeID:  "business_info_node_id",
					FieldID: "business_name",
					Alias:   "business_name",
				},
				{
					NodeID:  "bank_account_node_id",
					FieldID: "bank_name",
					Alias:   "bank_name",
				},
			},
			Condition: types.CrossNodeCondition{
				Type:     "custom_logic",
				Operator: "custom",
				Logic:    "business_name_matches_bank_name",
			},
			ErrorMsg:     "Business name should match or be consistent with the bank name",
			BusinessType: "", // Applies to all business types
			Severity:     types.ValidationSeverityWarning,
			Enabled:      true,
		},
		{
			ID:          "pan_signatory_consistency",
			Name:        "PAN and Signatory PAN Consistency",
			Description: "Ensures that the main PAN number matches the authorised signatory PAN for individual businesses",
			Fields: []types.CrossNodeFieldReference{
				{
					NodeID:  "pan_number_node_id",
					FieldID: "pan_number",
					Alias:   "pan_number",
				},
				{
					NodeID:  "authorised_signatory_node_id",
					FieldID: "signatory_pan",
					Alias:   "signatory_pan",
				},
			},
			Condition: types.CrossNodeCondition{
				Type:     "custom_logic",
				Operator: "custom",
				Logic:    "pan_matches_signatory_pan",
			},
			ErrorMsg:     "PAN number should match the authorised signatory PAN for individual businesses",
			BusinessType: "individual", // Only for individual business type
			Severity:     types.ValidationSeverityError,
			Enabled:      true,
		},
		{
			ID:          "address_completeness",
			Name:        "Address Information Completeness",
			Description: "Ensures that all address fields are properly filled",
			Fields: []types.CrossNodeFieldReference{
				{
					NodeID:  "business_info_node_id",
					FieldID: "business_city",
					Alias:   "business_city",
				},
				{
					NodeID:  "business_info_node_id",
					FieldID: "business_state",
					Alias:   "business_state",
				},
				{
					NodeID:  "business_info_node_id",
					FieldID: "business_pincode",
					Alias:   "business_pincode",
				},
			},
			Condition: types.CrossNodeCondition{
				Type:     "custom_logic",
				Operator: "custom",
				Logic:    "address_consistency",
			},
			ErrorMsg:     "All address fields (city, state, pincode) must be completed",
			BusinessType: "", // Applies to all business types
			Severity:     types.ValidationSeverityError,
			Enabled:      true,
		},
		{
			ID:          "payment_channel_url_consistency",
			Name:        "Payment Channel URL Consistency",
			Description: "Ensures that the selected payment channel has the appropriate URL fields filled",
			Fields: []types.CrossNodeFieldReference{
				{
					NodeID:  "payment_channel_node_id",
					FieldID: "payment_channel",
					Alias:   "payment_channel",
				},
				{
					NodeID:  "payment_channel_node_id",
					FieldID: "website_url",
					Alias:   "website_url",
				},
				{
					NodeID:  "payment_channel_node_id",
					FieldID: "android_url",
					Alias:   "android_url",
				},
				{
					NodeID:  "payment_channel_node_id",
					FieldID: "ios_url",
					Alias:   "ios_url",
				},
			},
			Condition: types.CrossNodeCondition{
				Type:     "field_contains",
				Operator: "contains",
				Fields:   []string{"payment_channel"},
				Value:    "website",
			},
			ErrorMsg:     "Website URL is required when payment channel is set to 'website'",
			BusinessType: "", // Applies to all business types
			Severity:     types.ValidationSeverityError,
			Enabled:      true,
		},
		{
			ID:          "business_document_consistency",
			Name:        "Business Document Consistency",
			Description: "Ensures that the required business documents are uploaded based on business type",
			Fields: []types.CrossNodeFieldReference{
				{
					NodeID:  "business_type_selection",
					FieldID: "business_type",
					Alias:   "business_type",
				},
				{
					NodeID:  "business_document_node_id",
					FieldID: "msme_document",
					Alias:   "msme_document",
				},
				{
					NodeID:  "business_document_node_id",
					FieldID: "certificate_of_incorporation",
					Alias:   "certificate_of_incorporation",
				},
				{
					NodeID:  "business_document_node_id",
					FieldID: "partnership_deed",
					Alias:   "partnership_deed",
				},
			},
			Condition: types.CrossNodeCondition{
				Type:     "field_match",
				Operator: "eq",
				Fields:   []string{"business_type"},
				Value:    "proprietorship",
			},
			ErrorMsg:     "MSME document is required for proprietorship business type",
			BusinessType: "proprietorship",
			Severity:     types.ValidationSeverityError,
			Enabled:      true,
		},
	}
}
