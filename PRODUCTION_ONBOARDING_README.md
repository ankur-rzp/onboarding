# Production Onboarding System

## Overview

The Production Onboarding System is a comprehensive, dynamic onboarding flow that supports all business types with conditional requirements based on the business type selected. It's designed to handle real-world onboarding scenarios with proper validation and document requirements.

## Features

### 🏢 Business Types Supported
- **Individual** - Personal PAN and basic documents
- **Proprietorship** - Personal PAN + MSME document
- **Private Limited** - Business PAN + CIN + Certificate of Incorporation
- **Public Limited** - Business PAN + CIN + Certificate of Incorporation
- **Partnership** - Business PAN + Partnership Deed
- **LLP** - Business PAN + Certificate of Incorporation
- **Trust** - Business PAN + Trust Deed
- **Society** - Business PAN + Society Registration Certificate
- **HUF** - Business PAN + HUF Deed

### 📋 Onboarding Flow
1. **Business Type Selection** - Choose your business type
2. **PAN Number** - Enter PAN details and upload document
3. **Payment Channel** - Select payment channel and optional URLs
4. **MCC & Policy Verification** - Business category and policy pages
5. **Business Document** - Upload required documents based on business type
6. **BMC Document** - Optional BMC document (subcategory dependent)
7. **Authorised Signatory Details** - Signatory information and documents
8. **Bank Account Details** - Bank account information with penny testing
9. **Business Information** - Business name, address, and contact details
10. **Onboarding Complete** - Final completion

### 🔧 Technical Features
- **Conditional Validation** - Documents are required/optional based on business type
- **File Upload Support** - Upload documents with progress tracking
- **Real-time Validation** - PAN, GST, CIN, and other field validation
- **Dynamic UI** - Adapts based on business type selection
- **Progress Tracking** - Visual progress indicator
- **Error Handling** - Comprehensive error messages and validation

## Getting Started

### Prerequisites
- Go 1.19 or higher
- Modern web browser

### Running the System

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Open the UI:**
   ```
   http://localhost:8080/dynamic-onboarding-ui.html
   ```

3. **Select the Production Onboarding graph** from the available options

### API Usage

#### List Available Graphs
```bash
curl http://localhost:8080/api/v1/graphs
```

#### Start a Session
```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"user_id": "demo-user", "graph_id": "PRODUCTION_GRAPH_ID"}'
```

#### Get Current Node
```bash
curl http://localhost:8080/api/v1/sessions/SESSION_ID/current
```

#### Submit Data
```bash
curl -X POST http://localhost:8080/api/v1/sessions/SESSION_ID/submit \
  -H "Content-Type: application/json" \
  -d '{"business_type": "private_limited"}'
```

## Demo Script

Run the demo script to see the system in action:

```bash
./demo_production_onboarding.sh
```

This script will:
1. List available graphs
2. Start a new session
3. Show the business type selection
4. Submit a business type
5. Display the next steps

## Business Type Requirements

### Individual
- Personal PAN number and document
- Basic contact information
- Bank account details
- Authorised signatory details

### Proprietorship
- Personal PAN number and document
- MSME document (required)
- GST document (optional)
- All other standard requirements

### Private/Public Limited
- Business PAN number and document
- CIN document (required)
- Certificate of Incorporation (required)
- GST document (optional)
- All other standard requirements

### Partnership
- Business PAN number and document
- Partnership Deed (required)
- GST document (optional)
- All other standard requirements

### LLP
- Business PAN number and document
- Certificate of Incorporation (required)
- GST document (optional)
- All other standard requirements

### Trust
- Business PAN number and document
- Trust Deed (required)
- GST document (optional)
- All other standard requirements

### Society
- Business PAN number and document
- Society Registration Certificate (required)
- GST document (optional)
- All other standard requirements

### HUF
- Business PAN number and document
- HUF Deed (required)
- GST document (optional)
- All other standard requirements

## Validation Rules

### PAN Number
- Format: `^[A-Z]{5}[0-9]{4}[A-Z]{1}$`
- Example: `ABCDE1234F`

### GST Number
- Format: `^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`
- Example: `22ABCDE1234F1Z5`

### CIN Number
- Format: `^[A-Z]{2}[0-9]{2}[A-Z]{2}[0-9]{4}$`
- Example: `U74999DL2014PTC272510`

### Bank Account
- Account Number: 9-18 digits
- IFSC Code: `^[A-Z]{4}0[A-Z0-9]{6}$`
- Example: `SBIN0001234`

### Phone Number
- Format: `^[6-9][0-9]{9}$`
- Example: `9876543210`

### Pincode
- Format: `^[1-9][0-9]{5}$`
- Example: `110001`

## File Upload

The system supports file uploads with:
- Progress tracking
- File validation
- Multiple file types
- Secure storage

### Supported File Types
- PDF documents
- Image files (JPG, PNG)
- Other document formats

## Error Handling

The system provides comprehensive error handling:
- Field validation errors
- File upload errors
- Business logic errors
- Network errors

## Testing

Run the test suite:

```bash
go test ./examples/ -v
```

This will test:
- Graph structure integrity
- Business type requirements mapping
- Linear flow validation
- Conditional validation rules
- Field validation patterns

## Architecture

### Components
- **Graph Engine** - Manages the onboarding flow
- **Validation Engine** - Handles field and business logic validation
- **File Upload Service** - Manages document uploads
- **API Layer** - RESTful API for frontend integration
- **UI Layer** - Dynamic web interface

### Data Flow
1. User selects business type
2. System determines required fields and documents
3. User fills in information step by step
4. System validates data and documents
5. User progresses through the flow
6. System completes onboarding when all requirements are met

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.

