-- Rollback: Restore email + secondary_emails columns
-- Convert from normalized employee_emails table back to denormalized structure

BEGIN;

-- Re-add email and secondary_emails columns to employees table
ALTER TABLE employees ADD COLUMN email VARCHAR(255);
ALTER TABLE employees ADD COLUMN secondary_emails JSONB DEFAULT '[]'::jsonb;

-- Migrate data back: Set primary email (first email alphabetically)
UPDATE employees e
SET email = (
    SELECT email
    FROM employee_emails ee
    WHERE ee.employee_id = e.id
    ORDER BY ee.email
    LIMIT 1
);

-- Migrate data back: Set secondary emails (all except the first one)
UPDATE employees e
SET secondary_emails = (
    SELECT COALESCE(jsonb_agg(email ORDER BY email), '[]'::jsonb)
    FROM (
        SELECT email
        FROM employee_emails ee
        WHERE ee.employee_id = e.id
        ORDER BY ee.email
        OFFSET 1
    ) sub
);

-- Make email NOT NULL after data migration
ALTER TABLE employees ALTER COLUMN email SET NOT NULL;

-- Recreate the unique index on tenant_id + email
CREATE UNIQUE INDEX idx_tenant_email ON employees(tenant_id, email);

-- Drop employee_emails table
DROP TABLE IF EXISTS employee_emails;

COMMIT;

