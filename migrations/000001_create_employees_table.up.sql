-- Create employees table with multi-tenant support
CREATE TABLE IF NOT EXISTS employees (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    secondary_emails JSONB DEFAULT '[]'::jsonb,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create composite unique index for tenant_id + email
CREATE UNIQUE INDEX idx_tenant_email ON employees(tenant_id, email);

-- Create index on tenant_id for efficient queries
CREATE INDEX idx_tenant_id ON employees(tenant_id);

-- Create index on created_at for filtering
CREATE INDEX idx_created_at ON employees(created_at);

-- Add comment to table
COMMENT ON TABLE employees IS 'Multi-tenant employee records with complete data isolation';
COMMENT ON COLUMN employees.tenant_id IS 'Tenant identifier from JWT token';
COMMENT ON COLUMN employees.secondary_emails IS 'Array of secondary emails from merged employees';

