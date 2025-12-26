-- Migration: Normalize email schema
-- Move from email + secondary_emails JSONB to normalized employee_emails table
-- This allows efficient querying by any email and enforces uniqueness at DB level

BEGIN;

-- Create employee_emails table
CREATE TABLE employee_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_employee_emails_employee FOREIGN KEY (employee_id) 
        REFERENCES employees(id) ON DELETE CASCADE
);

-- Create unique index to ensure no duplicate emails within a tenant
CREATE UNIQUE INDEX idx_employee_emails_tenant_email ON employee_emails(tenant_id, email);

-- Create index on employee_id for efficient lookups
CREATE INDEX idx_employee_emails_employee_id ON employee_emails(employee_id);

-- Add comments
COMMENT ON TABLE employee_emails IS 'Normalized employee email addresses with tenant isolation';
COMMENT ON COLUMN employee_emails.employee_id IS 'Foreign key to employees table';
COMMENT ON COLUMN employee_emails.tenant_id IS 'Denormalized tenant_id for efficient querying';
COMMENT ON COLUMN employee_emails.email IS 'Email address - unique within tenant';

-- Migrate existing data: Insert primary emails
INSERT INTO employee_emails (employee_id, tenant_id, email, created_at)
SELECT id, tenant_id, email, created_at
FROM employees
WHERE email IS NOT NULL AND email != '';

-- Migrate existing data: Unpack secondary_emails JSONB array
INSERT INTO employee_emails (employee_id, tenant_id, email, created_at)
SELECT 
    e.id,
    e.tenant_id,
    jsonb_array_elements_text(e.secondary_emails),
    e.created_at
FROM employees e
WHERE e.secondary_emails IS NOT NULL 
  AND jsonb_array_length(e.secondary_emails) > 0;

-- Drop old columns from employees table
ALTER TABLE employees DROP COLUMN IF EXISTS email;
ALTER TABLE employees DROP COLUMN IF EXISTS secondary_emails;

-- Drop the old unique index on tenant_id + email (no longer needed)
DROP INDEX IF EXISTS idx_tenant_email;

COMMIT;

