# Dynamic Onboarding System - System Overview

## 🎯 Problem Solved

This system addresses the complex requirements of dynamic onboarding processes where:

- **Different client types** (companies, individuals, proprietors) require different information
- **Government regulations** change frequently and need to be dynamically applied
- **Users can go back** and modify data, requiring path recalculation
- **Fault tolerance** and **consistency** are critical for business operations
- **Scalability** and **maintainability** are essential

## 🏗️ Architecture

### Core Components

1. **Graph Engine** (`internal/onboarding/engine.go`)
   - Handles graph traversal and path determination
   - Validates data against node rules
   - Manages conditional logic for edge traversal

2. **Validation System** (`internal/onboarding/engine.go`)
   - Field-level validation (email, phone, PAN, Aadhaar, GST)
   - Custom rule validation for government compliance
   - Conditional validation based on previous inputs

3. **Storage Layer** (`internal/storage/storage.go`)
   - PostgreSQL for persistent data storage
   - Redis for session caching and performance
   - ACID transactions for data consistency

4. **API Layer** (`internal/api/handlers.go`)
   - RESTful API for all operations
   - Session management
   - Graph operations

5. **Configuration** (`internal/config/config.go`)
   - Environment-based configuration
   - Validation rules from YAML files
   - Government compliance rules

## 🔄 Dynamic Flow Example

### Company Onboarding Flow
```
Start → Company Type → Company Details → Director Info → Address → Tax Info → Bank → Documents → End
```

### Individual Onboarding Flow
```
Start → Personal Info → Contact → Identity → Employment → Bank → Documents → End
```

### Dynamic Path Determination
- **Conditional Fields**: If employment status is "employed", require company name and job title
- **Government Rules**: Different validation rules based on company type
- **Backward Navigation**: Users can go back and change data, paths recalculate automatically

## 🛡️ Fault Tolerance Features

1. **Retry Mechanism**
   - Configurable retry attempts (default: 3)
   - Exponential backoff between retries
   - Session state preservation during retries

2. **Error Handling**
   - Comprehensive validation with detailed error messages
   - Graceful degradation on service failures
   - Health check endpoints for monitoring

3. **Data Consistency**
   - ACID transactions in PostgreSQL
   - Session state synchronization
   - Audit trail of all changes

4. **Recovery**
   - Session resumption after failures
   - Data integrity checks
   - Automatic cleanup of expired sessions

## 📊 Key Features Demonstrated

### 1. Graph-Based Architecture
```go
// Create a graph with nodes and edges
graph := NewGraph("Company Onboarding", "Onboarding flow for companies")
startNode := NewNode(NodeTypeStart, "Company Type", "Select company type")
companyNode := NewNode(NodeTypeInput, "Company Details", "Enter company info")

// Define edge conditions
edge := NewEdge(startNode.ID, companyNode.ID, EdgeCondition{
    Type: "always", // or "field_value", "custom"
})
```

### 2. Dynamic Validation
```go
// Field validation with custom rules
field := Field{
    Name: "pan_number",
    Type: FieldTypeText,
    Validation: FieldValidation{
        Pattern: `^[A-Z]{5}[0-9]{4}[A-Z]{1}$`,
        CustomRules: []string{"pan_validation"},
    },
}
```

### 3. Government Compliance
```yaml
# config/validation_rules.yaml
rules:
  pan_validation:
    description: "Validates Indian PAN format and checksum"
    pattern: "^[A-Z]{5}[0-9]{4}[A-Z]{1}$"
  
  gst_validation:
    description: "Validates Indian GST number format"
    pattern: "^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$"
```

### 4. Session Management
```go
// Start a new session
session, err := service.StartSession(ctx, "user123", "company-graph")

// Submit data and get next step
result, err := service.SubmitNodeData(ctx, session.ID, data)

// Go back to previous step
previousNode, err := service.GoBack(ctx, session.ID)
```

## 🚀 Usage Examples

### Starting a Session
```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "graph_id": "company-onboarding-graph"
  }'
```

### Submitting Data
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/submit \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Example Corp",
    "registration_number": "MH1234567890",
    "incorporation_date": "2023-01-01",
    "business_category": "technology"
  }'
```

### Going Back
```bash
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/back
```

## 🔧 Configuration

### Environment Variables
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- `ONBOARDING_MAX_RETRIES`, `ONBOARDING_RETRY_DELAY`
- `ONBOARDING_SESSION_TIMEOUT`

### Validation Rules
- YAML-based configuration in `config/validation_rules.yaml`
- Custom validation functions
- Government compliance rules
- Dynamic field requirements

## 📈 Scalability Features

1. **Horizontal Scaling**
   - Stateless API design
   - Redis for session sharing
   - Database connection pooling

2. **Performance**
   - Redis caching for frequently accessed data
   - Efficient graph traversal algorithms
   - Optimized database queries

3. **Monitoring**
   - Health check endpoints
   - Comprehensive logging
   - Error tracking and metrics

## 🧪 Testing

The system includes comprehensive tests:
- Unit tests for validation engine
- Integration tests for API endpoints
- Graph traversal tests
- Custom rule validation tests

Run tests with:
```bash
go test ./...
```

## 🐳 Deployment

### Docker Compose
```bash
docker-compose up -d
```

### Manual Deployment
```bash
# Set up PostgreSQL and Redis
# Configure environment variables
# Run the application
go run main.go
```

## 🎉 Benefits Achieved

1. **Flexibility**: Easy to add new onboarding flows and validation rules
2. **Maintainability**: Clean separation of concerns and modular design
3. **Reliability**: Fault tolerance and error recovery mechanisms
4. **Scalability**: Designed for horizontal scaling and high performance
5. **Compliance**: Built-in support for government regulations
6. **User Experience**: Smooth navigation with back/forward capabilities

## 🔮 Future Enhancements

1. **Rule Engine**: Integration with external rule engines (OPA, Drools)
2. **Workflow Engine**: Integration with workflow engines for complex processes
3. **Analytics**: Dashboard for monitoring onboarding completion rates
4. **Multi-tenancy**: Support for multiple organizations
5. **API Gateway**: Integration with API gateways for enterprise deployment
6. **Machine Learning**: AI-powered validation and fraud detection

This system successfully demonstrates a production-ready, dynamic onboarding solution that addresses all the requirements for flexibility, durability, consistency, and fault tolerance.
