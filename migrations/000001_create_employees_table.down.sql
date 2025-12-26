-- Drop employees table and all associated indexes
DROP INDEX IF EXISTS idx_created_at;
DROP INDEX IF EXISTS idx_tenant_id;
DROP INDEX IF EXISTS idx_tenant_email;
DROP TABLE IF EXISTS employees;

