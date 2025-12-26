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

## Sharing Proto Definitions with Other Projects

This service exposes its event and API proto definitions as a Go module, allowing other projects to import and use the same types.

### For Consumers (Other Go Projects)

To use the event definitions or API types in your project:

```bash
# Install the module
go get github.com/cvele/employee-service@v1.0.0

# Or get the latest version
go get github.com/cvele/employee-service@latest
```

Then import in your Go code:

```go
import (
    eventsv1 "github.com/cvele/employee-service/api/events/v1"
    employeev1 "github.com/cvele/employee-service/api/employee/v1"
)

// Use the event types
func handleEvent(event *eventsv1.EmployeeCreatedEvent) {
    fmt.Printf("Employee created: %s\n", event.Event.Employee.Id)
}
```

### Available Packages

- `github.com/cvele/employee-service/api/events/v1` - Employee lifecycle events (Created, Updated, Deleted, Merged)
- `github.com/cvele/employee-service/api/employee/v1` - Employee service API definitions

**Note**: Replace `cvele` with the actual GitHub organization/username where this repository is hosted.

## Releases

### Creating a Release

Releases are automated via GitHub Actions and produce both Docker images and Go module versions.

#### Steps to Create a Release:

1. Go to the **Actions** tab in GitHub
2. Select the **Release** workflow
3. Click **Run workflow**
4. Enter the version following semantic versioning (e.g., `v1.0.0`, `v1.2.3`)
5. Select the branch to release from
6. Click **Run workflow**

#### What Happens During Release:

1. ✅ Version format validation
2. 🔄 Updates `Version` variable in `cmd/employee-service/main.go`
3. 📝 Commits the version change
4. 🏷️ Creates a Git tag
5. 🐳 Builds Docker image with version tag
6. 📤 Pushes to GitHub Container Registry (GHCR)
7. 📋 Creates GitHub Release with changelog
8. 🎯 Tags both Docker image and Go module with the same version

#### Release Outputs:

- **Git Tag**: `v1.0.0` (for Go modules)
- **Docker Image**: `ghcr.io/cvele/employee-service:v1.0.0`
- **Docker Latest**: `ghcr.io/cvele/employee-service:latest`
- **GitHub Release**: With full changelog and usage instructions

### Using Released Versions

**Docker:**
```bash
docker pull ghcr.io/cvele/employee-service:v1.0.0
docker run ghcr.io/cvele/employee-service:v1.0.0
```

**Go Module:**
```bash
go get github.com/cvele/employee-service@v1.0.0
```

### Version Management

- Follow [Semantic Versioning](https://semver.org/): `vMAJOR.MINOR.PATCH`
- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible
- First production release should be `v1.0.0`
- Development releases can use `v0.x.x`

## License

See LICENSE file.

