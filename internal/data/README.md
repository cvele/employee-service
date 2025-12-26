# Data Layer

The data layer is responsible for interacting with external data sources (databases, message queues, etc.) following the Repository pattern as recommended by Kratos framework.

## Structure

### Core Files

- **data.go**: Main data layer initialization and dependency injection setup
  - Defines the `Data` struct containing shared resources (database, NATS connection, event publisher)
  - Provides `NewData()` constructor with cleanup function
  - Exports `ProviderSet` for Wire dependency injection

### Employee Domain

- **employee_model.go**: GORM model definitions and conversions
  - `EmployeeModel`: GORM entity with database mappings
  - `StringArray`: Custom type for PostgreSQL JSONB array handling
  - Model conversion functions (`ToEntity`, `FromEntity`)

- **employee_repo.go**: Repository implementation
  - `employeeRepo`: Implements `biz.EmployeeRepo` interface
  - CRUD operations: Create, Update, Delete, GetByID, GetByEmail
  - Advanced operations: List with pagination, CheckEmailExists, MergeEmployees
  - Transaction handling for complex operations

### Event Publishing

- **event_publisher.go**: Event publishing abstraction
  - `EventPublisher`: Publishes domain events to NATS
  - Event types: Created, Updated, Deleted, Merged
  - Implements retry logic and error handling

- **event_publisher_test.go**: Event contract tests
  - Validates event structure and required fields
  - Ensures backward compatibility

## Best Practices

1. **Separation of Concerns**: Models and repository logic are separated into different files
2. **Repository Pattern**: All data access goes through repository interfaces defined in `biz` layer
3. **Multi-tenancy**: All operations are tenant-scoped for data isolation
4. **Transaction Support**: Complex operations use GORM transactions
5. **Event-Driven**: Domain events are published for all state changes
6. **Context Propagation**: All methods accept `context.Context` for cancellation and tracing