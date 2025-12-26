# Testing Scripts

This directory contains utility scripts for testing the Employee Service.

## Generate JWT Token

Generate a JWT token for testing:

```bash
go run scripts/generate-jwt.go <secret> <user_id> <tenant_id>
```

Example:
```bash
go run scripts/generate-jwt.go my-secret user-123 tenant-abc
```

This will output a JWT token that can be used in the `Authorization: Bearer <token>` header.

## API Test Script

Run automated API tests:

```bash
# Set JWT secret (or it will use default "test-secret-key")
export JWT_SECRET="your-secret-key"

# Run the test script
./scripts/test-api.sh
```

The test script will:
1. Generate JWT tokens for two different tenants
2. Create employees in both tenants
3. Test multi-tenant isolation
4. Test duplicate email validation
5. Test that same email can exist in different tenants
6. Test update operations
7. Test query by email

**Note**: The service must be running at http://localhost:8000 for the tests to work.

## Requirements

- Service running at http://localhost:8000
- PostgreSQL database configured and running
- `jq` command-line tool (for pretty-printing JSON)
- `curl` command-line tool

