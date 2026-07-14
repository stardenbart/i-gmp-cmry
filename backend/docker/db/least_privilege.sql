-- =========================================================
-- Least Privilege: Create dedicated application DB user
-- Run as PostgreSQL superuser/admin ONCE during initial setup
-- =========================================================

-- Create dedicated app user
CREATE USER monitoring_app WITH PASSWORD 'change_this_strong_password_in_production';

-- Grant connection to the database
GRANT CONNECT ON DATABASE monitoring_audit TO monitoring_app;

-- Grant usage on the public schema
GRANT USAGE ON SCHEMA public TO monitoring_app;

-- Grant only necessary privileges on existing tables
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO monitoring_app;

-- Grant usage on sequences (for auto-increment types if any)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO monitoring_app;

-- Ensure future tables also get proper privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA public 
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO monitoring_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public 
    GRANT USAGE, SELECT ON SEQUENCES TO monitoring_app;

-- IMPORTANT: Do NOT grant DROP, TRUNCATE, CREATE, ALTER, or SUPERUSER
-- Those should only be used during migrations with the admin account
