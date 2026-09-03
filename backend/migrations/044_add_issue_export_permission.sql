INSERT INTO "Permission_Master" ("PermissionID", "ModuleID", "PermissionCode", "PermissionName")
VALUES ('PERM-ISS-E', 'MOD-ISS', 'EXPORT', 'Export Issue Report')
ON CONFLICT ("PermissionID") DO UPDATE SET
    "ModuleID" = EXCLUDED."ModuleID",
    "PermissionCode" = EXCLUDED."PermissionCode",
    "PermissionName" = EXCLUDED."PermissionName";

INSERT INTO "Role_Permission"
    ("RolePermissionID", "RoleID", "PermissionID", "IsAllowed")
VALUES
    ('RP-ADM-ISS-E', 'ROLE-001', 'PERM-ISS-E', TRUE),
    ('RP-AUD-ISS-E', 'ROLE-002', 'PERM-ISS-E', TRUE),
    ('RP-MGR-ISS-E', 'ROLE-005', 'PERM-ISS-E', TRUE)
ON CONFLICT ("RoleID", "PermissionID") DO UPDATE SET
    "IsAllowed" = TRUE,
    "RolePermissionUpdatedAt" = CURRENT_TIMESTAMP;
