# Database Migrations

This directory contains SQL migration files for the Employee Service database schema.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

## Migration Files

Migrations are numbered sequentially:
- `000001_create_employees_table.up.sql` - Creates employees table with indexes
- `000001_create_employees_table.down.sql` - Drops employees table

### Naming Convention

```
{version}_{description}.up.sql    # Apply migration
{version}_{description}.down.sql  # Rollback migration
```

## Usage

### Using the migrate command

```bash
# Apply all migrations
go run cmd/migrate/main.go -command up

# Apply specific number of migrations
go run cmd/migrate/main.go -command up -steps 1

# Rollback all migrations
go run cmd/migrate/main.go -command down

# Rollback specific number of migrations
go run cmd/migrate/main.go -command down -steps 1

# Check current version
go run cmd/migrate/main.go -command version

# Force to a specific version (use carefully!)
go run cmd/migrate/main.go -command force -steps 1

# Drop all tables (DANGEROUS!)
go run cmd/migrate/main.go -command drop
```

### Custom Database URL

```bash
# Set via environment variable
export DATABASE_URL="postgres://user:password@host:port/dbname?sslmode=disable"
go run cmd/migrate/main.go -command up

# Or via flag
go run cmd/migrate/main.go \
  -database-url "postgres://user:password@host:port/dbname?sslmode=disable" \
  -command up
```

### Using Makefile (recommended)

```bash
# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-status

# Create new migration
make migrate-create name=add_column_to_employees
```

## Creating New Migrations

### Manual Creation

1. Create two files in the `migrations/` directory:
   - `{next_version}_{description}.up.sql` - Migration to apply
   - `{next_version}_{description}.down.sql` - Migration to rollback

Example:
```bash
# Create files
touch migrations/000002_add_phone_to_employees.up.sql
touch migrations/000002_add_phone_to_employees.down.sql

# Edit up migration
cat > migrations/000002_add_phone_to_employees.up.sql << 'EOF'
ALTER TABLE employees ADD COLUMN phone VARCHAR(50);
CREATE INDEX idx_phone ON employees(phone);
EOF

# Edit down migration
cat > migrations/000002_add_phone_to_employees.down.sql << 'EOF'
DROP INDEX IF EXISTS idx_phone;
ALTER TABLE employees DROP COLUMN IF EXISTS phone;
EOF
```

### Using migrate CLI tool

Install globally:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Create new migration:
```bash
migrate create -ext sql -dir migrations -seq add_phone_to_employees
```

## Migration Best Practices

1. **Always provide both up and down migrations**
   - Every change should be reversible
   - Test both directions before committing

2. **Keep migrations atomic**
   - One logical change per migration
   - Use transactions where appropriate

3. **Never modify existing migrations**
   - Once applied to production, migrations are immutable
   - Create new migrations to fix issues

4. **Test migrations**
   - Test on a copy of production data
   - Verify performance on large tables
   - Check indexes are created

5. **Add comments**
   - Document why the change is needed
   - Explain any complex logic

6. **Handle data migrations carefully**
   - Consider backfilling in batches
   - Have rollback strategy for data changes
   - Monitor performance impact

## Deployment Strategy

### Development
```bash
# Apply all pending migrations
make migrate-up
```

### Production
```bash
# 1. Backup database first
pg_dump employee_service > backup.sql

# 2. Apply migrations
DATABASE_URL="production-url" make migrate-up

# 3. If issues occur, rollback
DATABASE_URL="production-url" make migrate-down -steps 1
```

### CI/CD Integration

```yaml
# Example GitHub Actions
- name: Run migrations
  run: |
    export DATABASE_URL="${{ secrets.DATABASE_URL }}"
    go run cmd/migrate/main.go -command up
```

## Troubleshooting

### Dirty Database State

If a migration fails midway, the database might be in a "dirty" state:

```bash
# Check current version and dirty status
go run cmd/migrate/main.go -command version

# Force to last known good version
go run cmd/migrate/main.go -command force -steps <last_good_version>

# Then fix the migration and retry
go run cmd/migrate/main.go -command up
```

### Migration Failed

1. Check the error message
2. Fix the SQL in the migration file
3. If already partially applied, you may need to manually clean up
4. Force version back if needed
5. Rerun migration

## Schema Versioning

The `schema_migrations` table tracks applied migrations:

```sql
SELECT * FROM schema_migrations;
```

**Never** manually modify this table unless you know what you're doing!

