-- Development data initialization script
-- This script is automatically run when PostgreSQL container starts for the first time

-- Note: The migrations should already have created the schema
-- This script just adds some sample data for development

-- Wait for migrations to complete (if any)
-- Insert sample data only if tables exist

DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'employees') THEN
        -- Insert sample employees for tenant-a
        INSERT INTO employees (tenant_id, email, secondary_emails, first_name, last_name, created_at, updated_at)
        VALUES 
            ('tenant-a', 'alice@example.com', '[]'::jsonb, 'Alice', 'Smith', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
            ('tenant-a', 'bob@example.com', '["bob.smith@example.com"]'::jsonb, 'Bob', 'Johnson', NOW() - INTERVAL '20 days', NOW() - INTERVAL '15 days'),
            ('tenant-a', 'charlie@example.com', '[]'::jsonb, 'Charlie', 'Brown', NOW() - INTERVAL '10 days', NOW() - INTERVAL '5 days')
        ON CONFLICT (tenant_id, email) DO NOTHING;

        -- Insert sample employees for tenant-b
        INSERT INTO employees (tenant_id, email, secondary_emails, first_name, last_name, created_at, updated_at)
        VALUES 
            ('tenant-b', 'david@example.com', '[]'::jsonb, 'David', 'Wilson', NOW() - INTERVAL '25 days', NOW() - INTERVAL '25 days'),
            ('tenant-b', 'eve@example.com', '["eve.anderson@example.com"]'::jsonb, 'Eve', 'Anderson', NOW() - INTERVAL '15 days', NOW() - INTERVAL '10 days'),
            ('tenant-b', 'alice@example.com', '[]'::jsonb, 'Alice', 'Davis', NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 days')
        ON CONFLICT (tenant_id, email) DO NOTHING;

        RAISE NOTICE 'Sample development data inserted successfully';
    ELSE
        RAISE NOTICE 'Employees table does not exist yet. Run migrations first.';
    END IF;
END $$;

