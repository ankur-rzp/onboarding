package examples

import (
	"onboarding-system/internal/types"
)

// CreateUnifiedOnboardingGraph creates a single graph that handles both individual and company onboarding
// with dynamic path activation based on user choices
func CreateUnifiedOnboardingGraph() *types.Graph {
	graph := types.NewGraph("Unified Onboarding", "Dynamic onboarding flow that adapts based on user type and choices")

	// Start node - User type selection
	startNode := types.NewNode(types.NodeTypeStart, "User Type Selection", "Select whether you're an individual or a company")
	startNode.Fields = []types.Field{
		{
			ID:       "user_type",
			Name:     "user_type",
			Type:     types.FieldTypeSelect,
			Required: true,
			Options:  []string{"individual", "company"},
			Validation: types.FieldValidation{
				CustomRules: []string{"user_type_validation"},
			},
		},
	}
	startNode.Validation = types.ValidationRules{
		RequiredFields: []string{"user_type"},
	}

	// Business Type Selection (for companies)
	businessTypeNode := types.NewNode(types.NodeTypeInput, "Business Type", "Select your business type")
	businessTypeNode.Fields = []types.Field{
		{
			ID:       "business_type",
			Name:     "business_type",
			Type:     types.FieldTypeSelect,
			Required: true,
			Options:  []string{"private_limited", "public_limited", "llp", "partnership", "sole_proprietorship"},
			Validation: types.FieldValidation{
				CustomRules: []string{"business_type_validation"},
			},
		},
	}
	businessTypeNode.Validation = types.ValidationRules{
		RequiredFields: []string{"business_type"},
	}

	// Personal Information (for individuals)
	personalInfoNode := types.NewNode(types.NodeTypeInput, "Personal Information", "Enter your personal details")
	personalInfoNode.Fields = []types.Field{
		{
			ID:       "first_name",
			Name:     "first_name",
			Type:     types.FieldTypeText,
			Required: true,
			Validation: types.FieldValidation{
				MinLength: 2,
				MaxLength: 50,
			},
		},
		{
			ID:       "last_name",
			Name:     "last_name",
			Type:     types.FieldTypeText,
			Required: true,
			Validation: types.FieldValidation{
				MinLength: 2,
				MaxLength: 50,
			},
		},
		{
			ID:       "date_of_birth",
			Name:     "date_of_birth",
			Type:     types.FieldTypeDate,
			Required: true,
		},
		{
			ID:       "gender",
			Name:     "gender",
			Type:     types.FieldTypeSelect,
			Required: true,
			Options:  []string{"male", "female", "other"},
		},
	}
	personalInfoNode.Validation = types.ValidationRules{
		RequiredFields: []string{"first_name", "last_name", "date_of_birth", "gender"},
	}

	// Company Information (for companies)
	companyInfoNode := types.NewNode(types.NodeTypeInput, "Company Information", "Enter your company details")
	companyInfoNode.Fields = []types.Field{
		{
			ID:       "company_name",
			Name:     "company_name",
			Type:     types.FieldTypeText,
			Required: true,
			Validation: types.FieldValidation{
				MinLength: 2,
				MaxLength: 100,
			},
		},
		{
			ID:       "registration_number",
			Name:     "registration_number",
			Type:     types.FieldTypeText,
			Required: false, // Only required for certain business types
			Validation: types.FieldValidation{
				Pattern:     `^[A-Z]{2}[0-9]{2}[A-Z]{2}[0-9]{4}$`,
				CustomRules: []string{"cin_validation"},
			},
		},
		{
			ID:       "incorporation_date",
			Name:     "incorporation_date",
			Type:     types.FieldTypeDate,
			Required: false, // Only required for certain business types
		},
	}
	companyInfoNode.Validation = types.ValidationRules{
		RequiredFields: []string{"company_name"},
		Conditions: []types.ValidationCondition{
			{
				Field:    "business_type",
				Operator: "in",
				Value:    []string{"private_limited", "public_limited", "llp"},
				Rule:     "registration_number and incorporation_date are required for registered companies",
			},
		},
	}

	// Contact Information (common for both)
	contactInfoNode := types.NewNode(types.NodeTypeInput, "Contact Information", "Enter your contact details")
	contactInfoNode.Fields = []types.Field{
		{
			ID:       "email",
			Name:     "email",
			Type:     types.FieldTypeEmail,
			Required: true,
		},
		{
			ID:       "phone",
			Name:     "phone",
			Type:     types.FieldTypeText,
			Required: true,
			Validation: types.FieldValidation{
				Pattern: `^[6-9][0-9]{9}$`,
			},
		},
		{
			ID:       "address_line1",
			Name:     "address_line1",
			Type:     types.FieldTypeText,
			Required: true,
			Validation: types.FieldValidation{
				MinLength: 5,
				MaxLength: 200,
			},
		},
		{
			ID:       "city",
			Name:     "city",
			Type:     types.FieldTypeText,
			Required: true,
			Validation: types.FieldValidation{
				MinLength: 2,
				MaxLength: 50,
			},
		},
		{
			ID:       "state",
			Name:     "state",
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
			ID:       "pincode",
			Name:     "pincode",
			Type:     types.FieldTypeText,
			Required: true,
			Validation: types.FieldValidation{
				Pattern: `^[1-9][0-9]{5}$`,
			},
		},
	}
	contactInfoNode.Validation = types.ValidationRules{
		RequiredFields: []string{"email", "phone", "address_line1", "city", "state", "pincode"},
	}

	// Identity Documents (common for both)
	identityNode := types.NewNode(types.NodeTypeInput, "Identity Documents", "Enter your identity document details")
	identityNode.Fields = []types.Field{
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
			ID:       "aadhaar_number",
			Name:     "aadhaar_number",
			Type:     types.FieldTypeText,
			Required: false, // Optional for companies
			Validation: types.FieldValidation{
				Pattern:     `^[0-9]{12}$`,
				CustomRules: []string{"aadhaar_validation"},
			},
		},
	}
	identityNode.Validation = types.ValidationRules{
		RequiredFields: []string{"pan_number"},
		Conditions: []types.ValidationCondition{
			{
				Field:    "user_type",
				Operator: "eq",
				Value:    "individual",
				Rule:     "aadhaar_number is required for individuals",
			},
		},
	}

	// Tax Information (conditional based on business type)
	taxInfoNode := types.NewNode(types.NodeTypeInput, "Tax Information", "Enter your tax registration details")
	taxInfoNode.Fields = []types.Field{
		{
			ID:       "gst_number",
			Name:     "gst_number",
			Type:     types.FieldTypeText,
			Required: false, // Only required for businesses with GST registration
			Validation: types.FieldValidation{
				Pattern:     `^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`,
				CustomRules: []string{"gst_validation"},
			},
		},
		{
			ID:       "tan_number",
			Name:     "tan_number",
			Type:     types.FieldTypeText,
			Required: false, // Only required for companies
			Validation: types.FieldValidation{
				Pattern: `^[A-Z]{4}[0-9]{5}[A-Z]{1}$`,
			},
		},
	}
	taxInfoNode.Validation = types.ValidationRules{
		Conditions: []types.ValidationCondition{
			{
				Field:    "business_type",
				Operator: "in",
				Value:    []string{"private_limited", "public_limited", "llp", "partnership"},
				Rule:     "gst_number is required for registered businesses",
			},
			{
				Field:    "user_type",
				Operator: "eq",
				Value:    "company",
				Rule:     "tan_number is required for companies",
			},
		},
	}

	// Bank Details (common for both)
	bankDetailsNode := types.NewNode(types.NodeTypeInput, "Bank Details", "Enter your bank account information")
	bankDetailsNode.Fields = []types.Field{
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
		{
			ID:       "account_number",
			Name:     "account_number",
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
			ID:       "account_type",
			Name:     "account_type",
			Type:     types.FieldTypeSelect,
			Required: true,
			Options:  []string{"savings", "current", "salary"},
		},
	}
	bankDetailsNode.Validation = types.ValidationRules{
		RequiredFields: []string{"bank_name", "account_number", "ifsc_code", "account_type"},
	}

	// Document Upload (conditional based on user type and business type)
	documentUploadNode := types.NewNode(types.NodeTypeInput, "Document Upload", "Upload required documents")
	documentUploadNode.Fields = []types.Field{
		{
			ID:       "pan_card",
			Name:     "pan_card",
			Type:     types.FieldTypeFile,
			Required: true,
		},
		{
			ID:       "address_proof",
			Name:     "address_proof",
			Type:     types.FieldTypeFile,
			Required: true,
		},
		{
			ID:       "certificate_of_incorporation",
			Name:     "certificate_of_incorporation",
			Type:     types.FieldTypeFile,
			Required: false, // Only for registered companies
		},
		{
			ID:       "gst_certificate",
			Name:     "gst_certificate",
			Type:     types.FieldTypeFile,
			Required: false, // Only for GST registered businesses
		},
	}
	documentUploadNode.Validation = types.ValidationRules{
		RequiredFields: []string{"pan_card", "address_proof"},
		Conditions: []types.ValidationCondition{
			{
				Field:    "business_type",
				Operator: "in",
				Value:    []string{"private_limited", "public_limited", "llp"},
				Rule:     "certificate_of_incorporation is required for registered companies",
			},
			{
				Field:    "gst_number",
				Operator: "ne",
				Value:    "",
				Rule:     "gst_certificate is required when GST number is provided",
			},
		},
	}

	// Completion node
	completionNode := types.NewNode(types.NodeTypeEnd, "Onboarding Complete", "Your onboarding is complete")
	completionNode.Fields = []types.Field{}

	// Add nodes to graph
	graph.Nodes[startNode.ID] = startNode
	graph.Nodes[businessTypeNode.ID] = businessTypeNode
	graph.Nodes[personalInfoNode.ID] = personalInfoNode
	graph.Nodes[companyInfoNode.ID] = companyInfoNode
	graph.Nodes[contactInfoNode.ID] = contactInfoNode
	graph.Nodes[identityNode.ID] = identityNode
	graph.Nodes[taxInfoNode.ID] = taxInfoNode
	graph.Nodes[bankDetailsNode.ID] = bankDetailsNode
	graph.Nodes[documentUploadNode.ID] = documentUploadNode
	graph.Nodes[completionNode.ID] = completionNode

	// Create edges with dynamic conditions
	// Start -> Business Type (for companies)
	startToBusinessType := types.NewEdge(startNode.ID, businessTypeNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "user_type",
		Operator: "eq",
		Value:    "company",
	})
	graph.Edges[startToBusinessType.ID] = startToBusinessType

	// Start -> Personal Info (for individuals)
	startToPersonalInfo := types.NewEdge(startNode.ID, personalInfoNode.ID, types.EdgeCondition{
		Type:     "field_value",
		Field:    "user_type",
		Operator: "eq",
		Value:    "individual",
	})
	graph.Edges[startToPersonalInfo.ID] = startToPersonalInfo

	// Business Type -> Company Info (for companies)
	businessTypeToCompanyInfo := types.NewEdge(businessTypeNode.ID, companyInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[businessTypeToCompanyInfo.ID] = businessTypeToCompanyInfo

	// Personal Info -> Contact Info (for individuals)
	personalInfoToContactInfo := types.NewEdge(personalInfoNode.ID, contactInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[personalInfoToContactInfo.ID] = personalInfoToContactInfo

	// Company Info -> Contact Info (for companies)
	companyInfoToContactInfo := types.NewEdge(companyInfoNode.ID, contactInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[companyInfoToContactInfo.ID] = companyInfoToContactInfo

	// Contact Info -> Identity
	contactInfoToIdentity := types.NewEdge(contactInfoNode.ID, identityNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[contactInfoToIdentity.ID] = contactInfoToIdentity

	// Identity -> Tax Info
	identityToTaxInfo := types.NewEdge(identityNode.ID, taxInfoNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[identityToTaxInfo.ID] = identityToTaxInfo

	// Tax Info -> Bank Details
	taxInfoToBankDetails := types.NewEdge(taxInfoNode.ID, bankDetailsNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[taxInfoToBankDetails.ID] = taxInfoToBankDetails

	// Bank Details -> Document Upload
	bankDetailsToDocumentUpload := types.NewEdge(bankDetailsNode.ID, documentUploadNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[bankDetailsToDocumentUpload.ID] = bankDetailsToDocumentUpload

	// Document Upload -> Completion
	documentUploadToCompletion := types.NewEdge(documentUploadNode.ID, completionNode.ID, types.EdgeCondition{
		Type: "always",
	})
	graph.Edges[documentUploadToCompletion.ID] = documentUploadToCompletion

	// Set start node
	graph.StartNodeID = startNode.ID

	return graph
}
