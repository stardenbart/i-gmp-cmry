-- =========================================================
-- Migration 026: Alter RolePermissionID column length
-- =========================================================

ALTER TABLE IF EXISTS "Role_Permission" ALTER COLUMN "RolePermissionID" TYPE VARCHAR(50);
