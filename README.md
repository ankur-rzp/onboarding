# Dynamic Onboarding System

A flexible, graph-based onboarding system built in Go that provides dynamic, fault-tolerant, and consistent onboarding flows for different types of clients (companies, individuals, proprietors).

## Features

- **Graph-based Flow**: Each onboarding process is represented as a connected graph where nodes are steps and edges define transitions
- **Dynamic Validation**: Each node has its own validation rules that can be customized
- **Fault Tolerance**: Built-in retry mechanisms and error handling
- **Durability**: PostgreSQL for persistent storage with Redis caching
- **Consistency**: ACID transactions and data integrity
- **Government Compliance**: Built-in support for Indian government regulations (PAN, Aadhaar, GST, etc.)
- **Flexible Paths**: Users can go back and change data, with paths adapting accordingly
- **Real-time Validation**: Immediate feedback on data submission

## 🎮 Live Demos

Preview the interactive demos directly in your browser:

### Main Demos
- **[Interactive Fintech Onboarding Demo](index.html)** - Full interactive onboarding flow visualization
  - 🔗 [Live Preview](https://htmlpreview.github.io/?https://raw.githubusercontent.com/ankur-rzp/onboarding/main/index.html) - View with all dependencies loaded
- **[Dynamic Onboarding System Demo](dynamic-onboarding-demo.html)** - CSV-based business type requirements with observer pattern

### Test Interfaces
- **[Dynamic Navigation Test](test-dynamic-navigation.html)** - Test dynamic node navigation and status updates
- **[Edge Visualization](test-edge-visualization.html)** - Visualize graph structure and edge relationships
- **[Cross-Node Validation Test](test-cross-node-validation.html)** - Test cross-node validation rules

> **Note**: These demos require the backend server to be running. See [Quick Start](#quick-start) to start the server locally, or use the API endpoints directly.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP API      │    │  Onboarding     │    │   Storage       │
│   (REST)        │◄──►│   Service       │◄──►│   (PostgreSQL   │
│                 │    │                 │    │    + Redis)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Graph Engine  │    │   Validation    │    │   Configuration │
│   (Traversal)   │    │   Engine        │    │   (YAML)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### Option 1: Local Development (In-Memory Storage) - Easiest

1. Clone the repository:
```bash
git clone <repository-url>
cd onboarding-system
```

2. Install dependencies:
```bash
go mod download
```

3. Run with in-memory storage (no database required):
```bash
./run-local.sh
```

Or manually:
```bash
go run main.go
```

The system will automatically use in-memory storage when no database configuration is provided.

### Option 2: Using Docker Compose (PostgreSQL + Redis)

1. Clone the repository:
```bash
git clone <repository-url>
cd onboarding-system
```

2. Start the system:
```bash
docker-compose up -d
```

3. The system will be available at `http://localhost:8080`

### Option 3: Manual Setup with Database

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL and Redis:
```bash
# PostgreSQL
createdb onboarding

# Redis (if not using Docker)
redis-server
```

3. Set environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=onboarding
export REDIS_HOST=localhost
export REDIS_PORT=6379
```

4. Run the application:
```bash
go run main.go
```

## API Usage

### Starting an Onboarding Session

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "graph_id": "company-onboarding-graph"
  }'
```

### Getting Current Node

```bash
curl http://localhost:8080/api/v1/sessions/{session_id}/current
```

### Submitting Node Data

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

## Example Onboarding Flows

### Company Onboarding
1. Company Type Selection
2. Company Details (name, registration, incorporation date)
3. Director Information (PAN, Aadhaar, contact details)
4. Business Address
5. Tax Information (GST, PAN, TAN)
6. Bank Details
7. Document Upload
8. Completion

### Individual Onboarding
1. Personal Information
2. Contact Information
3. Identity Documents (PAN, Aadhaar)
4. Employment Information
5. Bank Details
6. Document Upload
7. Completion

## Storage Options

The system supports two storage backends:

### In-Memory Storage (Default for Local Development)
- **No configuration required** - automatically used when no database config is provided
- **Perfect for development and testing** - no external dependencies
- **Data is lost on restart** - not suitable for production
- **Thread-safe** - supports concurrent access

### PostgreSQL + Redis Storage (Production)
- **Persistent storage** - data survives restarts
- **High performance** - Redis caching for frequently accessed data
- **ACID transactions** - data consistency guaranteed
- **Scalable** - supports multiple application instances

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_ADDRESS` | HTTP server address | `:8080` | No |
| `DB_HOST` | PostgreSQL host | `` | No* |
| `DB_PORT` | PostgreSQL port | `5432` | No* |
| `DB_USER` | PostgreSQL user | `` | No* |
| `DB_PASSWORD` | PostgreSQL password | `` | No* |
| `DB_NAME` | Database name | `` | No* |
| `REDIS_HOST` | Redis host | `localhost` | No* |
| `REDIS_PORT` | Redis port | `6379` | No* |
| `ONBOARDING_MAX_RETRIES` | Maximum retry attempts | `3` | No |
| `ONBOARDING_RETRY_DELAY` | Delay between retries | `5s` | No |
| `ONBOARDING_SESSION_TIMEOUT` | Session timeout | `24h` | No |

*Required only for PostgreSQL + Redis storage. If not provided, in-memory storage is used automatically.

### Validation Rules

Validation rules are configured in `config/validation_rules.yaml` and include:

- **PAN Validation**: Indian PAN number format and checksum
- **Aadhaar Validation**: Indian Aadhaar number format
- **GST Validation**: Indian GST number format
- **CIN Validation**: Indian Company Identification Number
- **Government Compliance**: KYC, AML, and data retention policies

## Graph Structure

### Node Types
- `start`: Entry point of the flow
- `input`: Data collection step
- `validation`: Validation-only step
- `decision`: Branching point
- `end`: Completion point

### Edge Conditions
- `always`: Always traverse this edge
- `field_value`: Traverse based on field value
- `custom`: Traverse based on custom rule

### Example Graph Definition

```go
// Create a new graph
graph := onboarding.NewGraph("Company Onboarding", "Onboarding flow for companies")

// Add nodes
startNode := onboarding.NewNode(onboarding.NodeTypeStart, "Company Type", "Select company type")
companyNode := onboarding.NewNode(onboarding.NodeTypeInput, "Company Details", "Enter company information")

// Add edges
edge := onboarding.NewEdge(startNode.ID, companyNode.ID, onboarding.EdgeCondition{
    Type: "always",
})

// Add to graph
graph.Nodes[startNode.ID] = startNode
graph.Nodes[companyNode.ID] = companyNode
graph.Edges[edge.ID] = edge
graph.StartNodeID = startNode.ID
```

## Fault Tolerance Features

1. **Retry Mechanism**: Automatic retry with exponential backoff
2. **Session Recovery**: Sessions can be resumed after failures
3. **Data Validation**: Comprehensive validation at each step
4. **Error Handling**: Graceful error handling with detailed messages
5. **Health Checks**: Built-in health check endpoints
6. **Graceful Shutdown**: Proper cleanup on application shutdown

## Development

### Project Structure

```
├── main.go                 # Application entry point
├── internal/
│   ├── api/               # HTTP handlers
│   ├── config/            # Configuration management
│   ├── onboarding/        # Core onboarding logic
│   └── storage/           # Data persistence layer
├── examples/              # Example onboarding flows
├── config/                # Configuration files
├── docker-compose.yml     # Docker Compose setup
└── Dockerfile            # Docker configuration
```

### Adding New Validation Rules

1. Add the rule to `config/validation_rules.yaml`
2. Implement the validation logic in `internal/onboarding/engine.go`
3. Update the `validateCustomRule` function

### Creating New Onboarding Flows

1. Create a new file in `examples/`
2. Define the graph structure with nodes and edges
3. Add validation rules for each node
4. Test the flow using the API

## Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestValidation ./internal/onboarding
```

## Monitoring

The system provides health check endpoints:

- `GET /health` - Basic health check
- `GET /api/v1/sessions/{id}` - Session status
- `GET /api/v1/graphs` - Available graphs

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue in the repository.
