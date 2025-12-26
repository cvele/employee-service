# Employee Service

A multi-tenant Employee microservice built with the Kratos framework, providing REST API for employee management with JWT authentication and complete tenant isolation.

## Features

- ✅ Multi-tenant architecture with complete data isolation
- ✅ JWT-based authentication with tenant_id extraction
- ✅ RESTful API for employee CRUD operations
- ✅ Pagination and date-based filtering
- ✅ Employee merge functionality
- ✅ PostgreSQL with GORM and UUIDs
- ✅ Email uniqueness per tenant
- ✅ Comprehensive error handling
- ✅ Input validation (proto-level and business logic)
- ✅ NATS event publishing for employee lifecycle events
- ✅ Database migrations with golang-migrate
- ✅ Docker Compose setup

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL
- Make

### Setup

#### Quick Start (Automated)

```bash
# Then run the service
go run ./cmd/employee-service -conf ./configs
```

#### Manual Setup

```bash
# 1. Install dependencies and tools
make init

# 2. Create environment file
cp env.development .env
# Edit .env if needed

# 3. Start PostgreSQL with Docker
make docker-up

# 4. Run database migrations
make migrate-up

# 5. Generate proto files
make api
make config
make generate

# 6. Build
make build

# 7. Run service
./bin/employee-service -conf ./configs
```

### Alternative: Full Development Environment

```bash
# Create .env file
cp env.development .env

# Start PostgreSQL, Redis, and pgAdmin
make docker-dev

# Run migrations
make migrate-up

# Run service
go run ./cmd/employee-service -conf ./configs
```

## Documentation

- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Complete implementation guide, API documentation, and architecture details
- **[EVENTS.md](EVENTS.md)** - NATS event publishing documentation and consumer examples
- **[VALIDATION.md](VALIDATION.md)** - Input validation rules and testing
- **[PROTOVALIDATE_MIGRATION.md](PROTOVALIDATE_MIGRATION.md)** - Migration to buf.build/protovalidate
- **[UUID_MIGRATION.md](UUID_MIGRATION.md)** - UUID implementation details
- **[DOCKER.md](DOCKER.md)** - Docker and Docker Compose guide
- **[scripts/README.md](scripts/README.md)** - Testing scripts and utilities
- **[migrations/README.md](migrations/README.md)** - Database migration guide

## API Endpoints

All endpoints require JWT authentication with `sub` and `tenant_id` claims:

- `POST /api/v1/employees` - Create employee
- `GET /api/v1/employees/{id}` - Get employee by ID
- `GET /api/v1/employees?email={email}` - Get employee by email
- `GET /api/v1/employees/list` - List employees with pagination
- `PUT /api/v1/employees/{id}` - Update employee
- `DELETE /api/v1/employees/{id}` - Delete employee
- `POST /api/v1/employees/merge` - Merge employees

## Testing

```bash
# Generate test JWT token
go run scripts/generate-jwt.go my-secret user-123 tenant-abc

# Run automated API tests (requires service running)
export JWT_SECRET="my-secret"
./scripts/test-api.sh

# Run NATS event tests (requires NATS and service running)
./scripts/test-events.sh

# Run event consumer (to monitor events)
make consumer
```

## Configuration

Edit `configs/config.yaml` for server and database settings. JWT secret is read from `JWT_SECRET` environment variable.

## Project Structure

```
employee-service/
├── api/employee/v1/          # Proto definitions
├── cmd/employee-service/      # Main application
├── internal/
│   ├── biz/                  # Business logic
│   ├── data/                 # Data access layer
│   ├── server/               # Server & middleware
│   └── service/              # Service layer
├── configs/                  # Configuration files
└── scripts/                  # Testing utilities
```

## Development

```bash
# Generate proto files
make api

# Generate wire dependencies
make generate

# Build
make build
```

## License

See LICENSE file.

