INSERT INTO "Plant_Master" ("PlantID", "PlantName") VALUES ('PLT-TEST', 'Test Plant');
INSERT INTO "Area_Master" ("AreaID", "PlantID", "AreaName") VALUES ('AREA-TEST', 'PLT-TEST', 'Test Area');
INSERT INTO "Aspek_Master" ("AspekID", "AreaID", "AspekName") VALUES ('ASP-1', 'AREA-TEST', 'Kebersihan Area');

DO $$
BEGIN
    BEGIN
        INSERT INTO "Aspek_Master" ("AspekID", "AreaID", "AspekName")
        VALUES ('ASP-2', 'AREA-TEST', '  KEBERSIHAN   AREA  ');
        RAISE EXCEPTION 'normalized duplicate unexpectedly accepted';
    EXCEPTION WHEN unique_violation THEN
        NULL;
    END;
END $$;

DO $$
BEGIN
    IF master_clean_text('  Kebersihan   Area  ') <> 'Kebersihan Area' THEN
        RAISE EXCEPTION 'master_clean_text mismatch';
    END IF;
    IF master_normalize_text('  KEBERSIHAN   AREA  ') <> 'kebersihan area' THEN
        RAISE EXCEPTION 'master_normalize_text mismatch';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM "Permission_Master" WHERE "PermissionID" = 'PERM-MSTR-I') THEN
        RAISE EXCEPTION 'import permission missing';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM "Master_Import_Batch" WHERE FALSE) THEN
        -- The SELECT itself verifies that the audit table exists.
        NULL;
    END IF;
END $$;

SELECT 'master_import_migration_ok' AS result;
