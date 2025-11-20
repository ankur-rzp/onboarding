package examples

import (
	"onboarding-system/internal/types"
)

// CreateDynamicProductionOnboardingGraph creates a dynamic onboarding graph with flexible navigation
func CreateDynamicProductionOnboardingGraph() *types.Graph {
	graph := types.NewGraph("Dynamic Production Onboarding", "Dynamic onboarding flow with flexible navigation based on dependencies")

	// Create all nodes (same as before)
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
	}

	completionNode := &types.Node{
		ID:          "completion_node_id",
		Type:        "end",
		Name:        "Onboarding Complete",
		Description: "Your onboarding is complete and under review",
		Fields:      []types.Field{},
		Validation:  types.ValidationRules{},
	}

	// Add all nodes to the graph
	graph.Nodes[startNode.ID] = startNode
	graph.Nodes[panNode.ID] = panNode
	graph.Nodes[paymentChannelNode.ID] = paymentChannelNode
	graph.Nodes[mccPolicyNode.ID] = mccPolicyNode
	graph.Nodes[businessDocumentNode.ID] = businessDocumentNode
	graph.Nodes[bmcDocumentNode.ID] = bmcDocumentNode
	graph.Nodes[authorisedSignatoryNode.ID] = authorisedSignatoryNode
	graph.Nodes[bankAccountNode.ID] = bankAccountNode
	graph.Nodes[businessInfoNode.ID] = businessInfoNode
	graph.Nodes[completionNode.ID] = completionNode

	// Create dynamic edges with flexible navigation
	// Start node can go to independent nodes (PAN, Payment Channel, MCC & Policy, Business Info)
	startToPan := types.NewEdge(startNode.ID, panNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[startToPan.ID] = startToPan

	startToPaymentChannel := types.NewEdge(startNode.ID, paymentChannelNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[startToPaymentChannel.ID] = startToPaymentChannel

	startToMccPolicy := types.NewEdge(startNode.ID, mccPolicyNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[startToMccPolicy.ID] = startToMccPolicy

	startToBusinessInfo := types.NewEdge(startNode.ID, businessInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[startToBusinessInfo.ID] = startToBusinessInfo

	// Independent nodes can connect to each other
	panToPaymentChannel := types.NewEdge(panNode.ID, paymentChannelNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[panToPaymentChannel.ID] = panToPaymentChannel

	panToMccPolicy := types.NewEdge(panNode.ID, mccPolicyNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[panToMccPolicy.ID] = panToMccPolicy

	panToBusinessInfo := types.NewEdge(panNode.ID, businessInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[panToBusinessInfo.ID] = panToBusinessInfo

	paymentChannelToMccPolicy := types.NewEdge(paymentChannelNode.ID, mccPolicyNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[paymentChannelToMccPolicy.ID] = paymentChannelToMccPolicy

	paymentChannelToBusinessInfo := types.NewEdge(paymentChannelNode.ID, businessInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[paymentChannelToBusinessInfo.ID] = paymentChannelToBusinessInfo

	mccPolicyToBusinessInfo := types.NewEdge(mccPolicyNode.ID, businessInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[mccPolicyToBusinessInfo.ID] = mccPolicyToBusinessInfo

	// Dependent nodes - only accessible when dependencies are satisfied
	// Business Document depends on business_type being selected
	businessDocumentFromStart := types.NewEdge(startNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[businessDocumentFromStart.ID] = businessDocumentFromStart

	businessDocumentFromPan := types.NewEdge(panNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[businessDocumentFromPan.ID] = businessDocumentFromPan

	businessDocumentFromPaymentChannel := types.NewEdge(paymentChannelNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[businessDocumentFromPaymentChannel.ID] = businessDocumentFromPaymentChannel

	businessDocumentFromMccPolicy := types.NewEdge(mccPolicyNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[businessDocumentFromMccPolicy.ID] = businessDocumentFromMccPolicy

	businessDocumentFromBusinessInfo := types.NewEdge(businessInfoNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[businessDocumentFromBusinessInfo.ID] = businessDocumentFromBusinessInfo

	// BMC Document depends on subcategory being selected
	bmcDocumentFromMccPolicy := types.NewEdge(mccPolicyNode.ID, bmcDocumentNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "subcategory",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[bmcDocumentFromMccPolicy.ID] = bmcDocumentFromMccPolicy

	bmcDocumentFromBusinessDocument := types.NewEdge(businessDocumentNode.ID, bmcDocumentNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[bmcDocumentFromBusinessDocument.ID] = bmcDocumentFromBusinessDocument

	// Authorised Signatory depends on business_type being selected
	authorisedSignatoryFromStart := types.NewEdge(startNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[authorisedSignatoryFromStart.ID] = authorisedSignatoryFromStart

	authorisedSignatoryFromPan := types.NewEdge(panNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[authorisedSignatoryFromPan.ID] = authorisedSignatoryFromPan

	authorisedSignatoryFromPaymentChannel := types.NewEdge(paymentChannelNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[authorisedSignatoryFromPaymentChannel.ID] = authorisedSignatoryFromPaymentChannel

	authorisedSignatoryFromMccPolicy := types.NewEdge(mccPolicyNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[authorisedSignatoryFromMccPolicy.ID] = authorisedSignatoryFromMccPolicy

	authorisedSignatoryFromBusinessInfo := types.NewEdge(businessInfoNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[authorisedSignatoryFromBusinessInfo.ID] = authorisedSignatoryFromBusinessInfo

	authorisedSignatoryFromBusinessDocument := types.NewEdge(businessDocumentNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[authorisedSignatoryFromBusinessDocument.ID] = authorisedSignatoryFromBusinessDocument

	authorisedSignatoryFromBmcDocument := types.NewEdge(bmcDocumentNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[authorisedSignatoryFromBmcDocument.ID] = authorisedSignatoryFromBmcDocument

	// Bank Account depends on business_type being selected
	bankAccountFromStart := types.NewEdge(startNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[bankAccountFromStart.ID] = bankAccountFromStart

	bankAccountFromPan := types.NewEdge(panNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[bankAccountFromPan.ID] = bankAccountFromPan

	bankAccountFromPaymentChannel := types.NewEdge(paymentChannelNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[bankAccountFromPaymentChannel.ID] = bankAccountFromPaymentChannel

	bankAccountFromMccPolicy := types.NewEdge(mccPolicyNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[bankAccountFromMccPolicy.ID] = bankAccountFromMccPolicy

	bankAccountFromBusinessInfo := types.NewEdge(businessInfoNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "business_type",
		Operator: "ne",
		Value:    "",
	})
	graph.Edges[bankAccountFromBusinessInfo.ID] = bankAccountFromBusinessInfo

	bankAccountFromAuthorisedSignatory := types.NewEdge(authorisedSignatoryNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[bankAccountFromAuthorisedSignatory.ID] = bankAccountFromAuthorisedSignatory

	// Completion node - accessible from any node when all requirements are met
	// This will be handled by the dynamic engine based on rule groups
	completionFromStart := types.NewEdge(startNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromStart.ID] = completionFromStart

	completionFromPan := types.NewEdge(panNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromPan.ID] = completionFromPan

	completionFromPaymentChannel := types.NewEdge(paymentChannelNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromPaymentChannel.ID] = completionFromPaymentChannel

	completionFromMccPolicy := types.NewEdge(mccPolicyNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromMccPolicy.ID] = completionFromMccPolicy

	completionFromBusinessDocument := types.NewEdge(businessDocumentNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromBusinessDocument.ID] = completionFromBusinessDocument

	completionFromBmcDocument := types.NewEdge(bmcDocumentNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromBmcDocument.ID] = completionFromBmcDocument

	completionFromAuthorisedSignatory := types.NewEdge(authorisedSignatoryNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromAuthorisedSignatory.ID] = completionFromAuthorisedSignatory

	completionFromBankAccount := types.NewEdge(bankAccountNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromBankAccount.ID] = completionFromBankAccount

	completionFromBusinessInfo := types.NewEdge(businessInfoNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "completion_check",
	})
	graph.Edges[completionFromBusinessInfo.ID] = completionFromBusinessInfo

	// Set the start node
	graph.StartNodeID = startNode.ID

	return graph
}
