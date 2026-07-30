-- Migration 016: Sync LastInspection for existing completed inspection headers
UPDATE "DetailKawasan_Master" dk
SET "LastInspection" = (
    SELECT MAX(h."InspectionHeaderCreatedAt")
    FROM "Inspection_Header" h
    WHERE h."DetailKawasanID" = dk."DetailKawasanID"
      AND h."InspectionHeaderStatus" IN ('Completed', 'Approved')
)
WHERE EXISTS (
    SELECT 1 FROM "Inspection_Header" h
    WHERE h."DetailKawasanID" = dk."DetailKawasanID"
      AND h."InspectionHeaderStatus" IN ('Completed', 'Approved')
);

UPDATE "Kawasan_Master" k
SET "LastInspection" = (
    SELECT MAX(h."InspectionHeaderCreatedAt")
    FROM "Inspection_Header" h
    WHERE h."KawasanID" = k."KawasanID"
      AND h."InspectionHeaderStatus" IN ('Completed', 'Approved')
)
WHERE EXISTS (
    SELECT 1 FROM "Inspection_Header" h
    WHERE h."KawasanID" = k."KawasanID"
      AND h."InspectionHeaderStatus" IN ('Completed', 'Approved')
);
