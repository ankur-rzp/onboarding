package examples

import (
	"onboarding-system/internal/types"
)

// CreateProductionOnboardingGraph creates a comprehensive graph that handles all business types
// with dynamic path activation based on business type selection and conditional requirements
func CreateProductionOnboardingGraph() *types.Graph {
	graph := types.NewGraph("Production Onboarding", "Comprehensive onboarding flow for all business types with conditional requirements")

	// Start node - Business type selection
	startNode := types.NewNode(types.NodeTypeStart, "Business Type Selection", "Select your business type")
	startNode.Fields = []types.Field{
		{
			ID:       "business_type",
			Name:     "business_type",
			Type:     types.FieldTypeSelect,
			Required: true,
			Options: []string{
				"individual", "proprietorship", "private_limited", "public_limited",
				"partnership", "llp", "trust", "society", "huf",
			},
			Validation: types.FieldValidation{
				CustomRules: []string{"business_type_validation"},
			},
		},
	}
	startNode.Validation = types.ValidationRules{
		RequiredFields: []string{"business_type"},
	}

	// PAN Number node - Required for all business types
	panNode := types.NewNode(types.NodeTypeInput, "PAN Number", "Enter your PAN number")
	panNode.Fields = []types.Field{
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
	}
	panNode.Validation = types.ValidationRules{
		RequiredFields: []string{"pan_number", "pan_document"},
	}

	// Payment Channel node - Required for all business types
	paymentChannelNode := types.NewNode(types.NodeTypeInput, "Payment Channel", "Select your payment channel")
	paymentChannelNode.Fields = []types.Field{
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
			Required: false, // Conditional based on payment_channel selection
			Validation: types.FieldValidation{
				Pattern: `^https?://[^\s/$.?#].[^\s]*$`,
			},
		},
		{
			ID:       "android_url",
			Name:     "android_url",
			Type:     types.FieldTypeText,
			Required: false, // Conditional based on payment_channel selection
			Validation: types.FieldValidation{
				Pattern: `^https?://[^\s/$.?#].[^\s]*$`,
			},
		},
		{
			ID:       "ios_url",
			Name:     "ios_url",
			Type:     types.FieldTypeText,
			Required: false, // Conditional based on payment_channel selection
			Validation: types.FieldValidation{
				Pattern: `^https?://[^\s/$.?#].[^\s]*$`,
			},
		},
	}
	paymentChannelNode.Validation = types.ValidationRules{
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
	}

	// MCC & Policy Verification node - Required for all business types
	mccPolicyNode := types.NewNode(types.NodeTypeInput, "MCC & Policy Verification", "Provide business category and policy information")
	mccPolicyNode.Fields = []types.Field{
		{
			ID:       "subcategory",
			Name:     "subcategory",
			Type:     types.FieldTypeSelect,
			Required: true,
			Options: []string{
				"retail", "wholesale", "services", "manufacturing", "technology",
				"healthcare", "education", "finance", "real_estate", "hospitality",
				"transportation", "agriculture", "consulting", "other",
			},
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
	}
	mccPolicyNode.Validation = types.ValidationRules{
		RequiredFields: []string{"subcategory", "policy_pages"},
	}

	// Business Document node - Conditional based on business type
	businessDocumentNode := types.NewNode(types.NodeTypeInput, "Business Document", "Upload required business documents")
	businessDocumentNode.Fields = []types.Field{
		{
			ID:       "msme_document",
			Name:     "msme_document",
			Type:     types.FieldTypeFile,
			Required: false, // Required for proprietorship
		},
		{
			ID:       "gst_document",
			Name:     "gst_document",
			Type:     types.FieldTypeFile,
			Required: false, // Optional for most business types
		},
		{
			ID:       "gstin_number",
			Name:     "gstin_number",
			Type:     types.FieldTypeText,
			Required: false, // Optional for most business types
			Validation: types.FieldValidation{
				Pattern:     `^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`,
				CustomRules: []string{"gst_validation"},
			},
		},
		{
			ID:       "cin_document",
			Name:     "cin_document",
			Type:     types.FieldTypeFile,
			Required: false, // Required for private_limited and public_limited
		},
		{
			ID:       "cin_number",
			Name:     "cin_number",
			Type:     types.FieldTypeText,
			Required: false, // Required for private_limited and public_limited
			Validation: types.FieldValidation{
				Pattern: `^[A-Z]{2}[0-9]{2}[A-Z]{2}[0-9]{4}$`,
			},
		},
		{
			ID:       "certificate_of_incorporation",
			Name:     "certificate_of_incorporation",
			Type:     types.FieldTypeFile,
			Required: false, // Required for private_limited, public_limited, and llp
		},
		{
			ID:       "partnership_deed",
			Name:     "partnership_deed",
			Type:     types.FieldTypeFile,
			Required: false, // Required for partnership
		},
		{
			ID:       "trust_deed",
			Name:     "trust_deed",
			Type:     types.FieldTypeFile,
			Required: false, // Required for trust
		},
		{
			ID:       "society_registration_certificate",
			Name:     "society_registration_certificate",
			Type:     types.FieldTypeFile,
			Required: false, // Required for society
		},
		{
			ID:       "huf_deed",
			Name:     "huf_deed",
			Type:     types.FieldTypeFile,
			Required: false, // Required for huf
		},
	}
	businessDocumentNode.Validation = types.ValidationRules{
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
				Value:    []string{"private_limited", "public_limited"},
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
	}

	// BMC Document node - Optional for all business types (depending on subcategory)
	bmcDocumentNode := types.NewNode(types.NodeTypeInput, "BMC Document", "Upload BMC document if required")
	bmcDocumentNode.Fields = []types.Field{
		{
			ID:       "bmc_document",
			Name:     "bmc_document",
			Type:     types.FieldTypeFile,
			Required: false, // Optional for all business types
		},
	}
	bmcDocumentNode.Validation = types.ValidationRules{
		RequiredFields: []string{}, // All fields are optional
	}

	// Authorised Signatory Details node - Required for all business types
	authorisedSignatoryNode := types.NewNode(types.NodeTypeInput, "Authorised Signatory Details", "Enter authorised signatory information")
	authorisedSignatoryNode.Fields = []types.Field{
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
	}
	authorisedSignatoryNode.Validation = types.ValidationRules{
		RequiredFields: []string{"signatory_name", "signatory_pan", "signatory_pan_document", "signatory_aadhaar_front", "signatory_aadhaar_back"},
	}

	// Bank Account node - Required for all business types
	bankAccountNode := types.NewNode(types.NodeTypeInput, "Bank Account Details", "Enter bank account information")
	bankAccountNode.Fields = []types.Field{
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
	}
	bankAccountNode.Validation = types.ValidationRules{
		RequiredFields: []string{"bank_account_number", "ifsc_code", "bank_name"},
		CustomRules:    []string{"penny_testing_verification"},
	}

	// Business Information node - Required for all business types
	businessInfoNode := types.NewNode(types.NodeTypeInput, "Business Information", "Enter business details")
	businessInfoNode.Fields = []types.Field{
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
			Options: []string{
				"andhra_pradesh", "arunachal_pradesh", "assam", "bihar", "chhattisgarh",
				"goa", "gujarat", "haryana", "himachal_pradesh", "jharkhand",
				"karnataka", "kerala", "madhya_pradesh", "maharashtra", "manipur",
				"meghalaya", "mizoram", "nagaland", "odisha", "punjab",
				"rajasthan", "sikkim", "tamil_nadu", "telangana", "tripura",
				"uttar_pradesh", "uttarakhand", "west_bengal",
			},
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
	}
	businessInfoNode.Validation = types.ValidationRules{
		RequiredFields: []string{"business_name", "brand_name", "business_address_line1", "business_city", "business_state", "business_pincode"},
	}

	// Completion node
	completionNode := types.NewNode(types.NodeTypeEnd, "Onboarding Complete", "Your onboarding is complete and under review")
	completionNode.Fields = []types.Field{}

	// Add nodes to graph
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

	// Create edges with sequential flow
	// Start -> PAN Number (always)
	startToPan := types.NewEdge(startNode.ID, panNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[startToPan.ID] = startToPan

	// PAN Number -> Payment Channel (always)
	panToPaymentChannel := types.NewEdge(panNode.ID, paymentChannelNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[panToPaymentChannel.ID] = panToPaymentChannel

	// Payment Channel -> MCC & Policy (always)
	paymentChannelToMccPolicy := types.NewEdge(paymentChannelNode.ID, mccPolicyNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[paymentChannelToMccPolicy.ID] = paymentChannelToMccPolicy

	// MCC & Policy -> Business Document (always)
	mccPolicyToBusinessDocument := types.NewEdge(mccPolicyNode.ID, businessDocumentNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[mccPolicyToBusinessDocument.ID] = mccPolicyToBusinessDocument

	// Business Document -> BMC Document (always - but BMC is optional)
	businessDocumentToBmcDocument := types.NewEdge(businessDocumentNode.ID, bmcDocumentNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[businessDocumentToBmcDocument.ID] = businessDocumentToBmcDocument

	// BMC Document -> Authorised Signatory (always)
	bmcDocumentToAuthorisedSignatory := types.NewEdge(bmcDocumentNode.ID, authorisedSignatoryNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[bmcDocumentToAuthorisedSignatory.ID] = bmcDocumentToAuthorisedSignatory

	// Authorised Signatory -> Bank Account (always)
	authorisedSignatoryToBankAccount := types.NewEdge(authorisedSignatoryNode.ID, bankAccountNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[authorisedSignatoryToBankAccount.ID] = authorisedSignatoryToBankAccount

	// Bank Account -> Business Information (always)
	bankAccountToBusinessInfo := types.NewEdge(bankAccountNode.ID, businessInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[bankAccountToBusinessInfo.ID] = bankAccountToBusinessInfo

	// Business Information -> Completion (always)
	businessInfoToCompletion := types.NewEdge(businessInfoNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[businessInfoToCompletion.ID] = businessInfoToCompletion

	// Set start node
	graph.StartNodeID = startNode.ID

	return graph
}

// GetBusinessTypeRequirements returns the specific requirements for each business type
func GetBusinessTypeRequirements() map[string]map[string]string {
	return map[string]map[string]string{
		"individual": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "optional",
			"certificate_of_incorporation":     "optional",
			"partnership_deed":                 "optional",
			"trust_deed":                       "optional",
			"society_registration_certificate": "optional",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"proprietorship": {
			"pan_document":                     "required",
			"msme_document":                    "required",
			"gst_document":                     "optional",
			"cin_document":                     "optional",
			"certificate_of_incorporation":     "optional",
			"partnership_deed":                 "optional",
			"trust_deed":                       "optional",
			"society_registration_certificate": "optional",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"private_limited": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "required",
			"certificate_of_incorporation":     "required",
			"partnership_deed":                 "optional",
			"trust_deed":                       "optional",
			"society_registration_certificate": "optional",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"public_limited": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "required",
			"certificate_of_incorporation":     "required",
			"partnership_deed":                 "optional",
			"trust_deed":                       "optional",
			"society_registration_certificate": "optional",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"partnership": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "optional",
			"certificate_of_incorporation":     "optional",
			"partnership_deed":                 "required",
			"trust_deed":                       "optional",
			"society_registration_certificate": "optional",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"llp": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "optional",
			"certificate_of_incorporation":     "required",
			"partnership_deed":                 "optional",
			"trust_deed":                       "optional",
			"society_registration_certificate": "optional",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"trust": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "optional",
			"certificate_of_incorporation":     "optional",
			"partnership_deed":                 "optional",
			"trust_deed":                       "required",
			"society_registration_certificate": "optional",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"society": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "optional",
			"certificate_of_incorporation":     "optional",
			"partnership_deed":                 "optional",
			"trust_deed":                       "optional",
			"society_registration_certificate": "required",
			"huf_deed":                         "optional",
			"bmc_document":                     "optional",
		},
		"huf": {
			"pan_document":                     "required",
			"msme_document":                    "optional",
			"gst_document":                     "optional",
			"cin_document":                     "optional",
			"certificate_of_incorporation":     "optional",
			"partnership_deed":                 "optional",
			"trust_deed":                       "optional",
			"society_registration_certificate": "optional",
			"huf_deed":                         "required",
			"bmc_document":                     "optional",
		},
	}
}
