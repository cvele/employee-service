-- Rollback: Change back to BIGSERIAL (data loss!)
BEGIN;

DROP TABLE IF EXISTS employees;

-- Recreate with BIGSERIAL
CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    secondary_emails JSONB DEFAULT '[]'::jsonb,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_tenant_email ON employees(tenant_id, email);
CREATE INDEX idx_tenant_id ON employees(tenant_id);
CREATE INDEX idx_created_at ON employees(created_at);

COMMIT;

