-- Menambahkan modul & permission untuk halaman Dashboard KPI & Analitik
-- (/kpi), supaya Admin bisa mengatur akses ke halaman ini per-role /
-- per-user dari layar "Roles" dan "Edit User" — sama seperti modul lain
-- (mis. Data Inspeksi (GMP)). MOD-KPI belum pernah ada di Module_Master
-- sebelumnya sehingga perlu di-insert juga, bukan cuma permission-nya.

INSERT INTO "Module_Master" ("ModuleID", "ModuleName")
VALUES ('MOD-KPI', 'Dashboard KPI & Analitik')
ON CONFLICT ("ModuleID") DO UPDATE SET
    "ModuleName" = EXCLUDED."ModuleName";

INSERT INTO "Permission_Master" ("PermissionID", "ModuleID", "PermissionCode", "PermissionName")
VALUES ('PERM-KPI-R', 'MOD-KPI', 'READ', 'View Dashboard KPI')
ON CONFLICT ("PermissionID") DO UPDATE SET
    "ModuleID" = EXCLUDED."ModuleID",
    "PermissionCode" = EXCLUDED."PermissionCode",
    "PermissionName" = EXCLUDED."PermissionName";

-- Admin-only secara default (belum pernah dibuka untuk role lain),
-- mengikuti pola PERM-GMP-R.
INSERT INTO "Role_Permission"
    ("RolePermissionID", "RoleID", "PermissionID", "IsAllowed")
VALUES
    ('RP-ADM-KPI-R', 'ROLE-001', 'PERM-KPI-R', TRUE)
ON CONFLICT ("RoleID", "PermissionID") DO UPDATE SET
    "IsAllowed" = TRUE,
    "RolePermissionUpdatedAt" = CURRENT_TIMESTAMP;
