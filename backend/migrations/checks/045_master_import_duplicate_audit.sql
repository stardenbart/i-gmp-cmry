-- Run this read-only audit before deploying migration 045 in an existing
-- environment. Any returned row must be reviewed; this script never deletes
-- or merges master data because those IDs may be referenced by inspections.

SELECT 'aspek' AS entity_type,
       "AreaID" AS parent_id,
       lower(regexp_replace(btrim(normalize("AspekName", NFC)), '[[:space:]]+', ' ', 'g')) AS normalized_value,
       array_agg("AspekID" ORDER BY "AspekID") AS conflicting_ids,
       count(*) AS duplicate_count
FROM "Aspek_Master"
GROUP BY "AreaID", lower(regexp_replace(btrim(normalize("AspekName", NFC)), '[[:space:]]+', ' ', 'g'))
HAVING count(*) > 1;

SELECT 'detail' AS entity_type,
       "AspekID" AS parent_id,
       lower(regexp_replace(btrim(normalize("DetailName", NFC)), '[[:space:]]+', ' ', 'g')) AS normalized_value,
       array_agg("DetailID" ORDER BY "DetailID") AS conflicting_ids,
       count(*) AS duplicate_count
FROM "Detail_Master"
GROUP BY "AspekID", lower(regexp_replace(btrim(normalize("DetailName", NFC)), '[[:space:]]+', ' ', 'g'))
HAVING count(*) > 1;

SELECT 'uraian' AS entity_type,
       "DetailID" AS parent_id,
       lower(regexp_replace(btrim(normalize("UraianText", NFC)), '[[:space:]]+', ' ', 'g')) AS normalized_value,
       array_agg("UraianID" ORDER BY "UraianID") AS conflicting_ids,
       count(*) AS duplicate_count
FROM "Uraian_Master"
GROUP BY "DetailID", lower(regexp_replace(btrim(normalize("UraianText", NFC)), '[[:space:]]+', ' ', 'g'))
HAVING count(*) > 1;

SELECT 'hei_name' AS entity_type,
       lower(regexp_replace(btrim(normalize("CategoryName", NFC)), '[[:space:]]+', ' ', 'g')) AS parent_id,
       lower(regexp_replace(btrim(normalize("HEIName", NFC)), '[[:space:]]+', ' ', 'g')) AS normalized_value,
       array_agg("HEIID" ORDER BY "HEIID") AS conflicting_ids,
       count(*) AS duplicate_count
FROM "HEI_Master"
GROUP BY lower(regexp_replace(btrim(normalize("CategoryName", NFC)), '[[:space:]]+', ' ', 'g')),
         lower(regexp_replace(btrim(normalize("HEIName", NFC)), '[[:space:]]+', ' ', 'g'))
HAVING count(*) > 1;

SELECT 'hei_code' AS entity_type,
       lower(regexp_replace(btrim(normalize("CategoryName", NFC)), '[[:space:]]+', ' ', 'g')) AS parent_id,
       upper(lower(regexp_replace(btrim(normalize("HEICode", NFC)), '[[:space:]]+', ' ', 'g'))) AS normalized_value,
       array_agg("HEIID" ORDER BY "HEIID") AS conflicting_ids,
       count(*) AS duplicate_count
FROM "HEI_Master"
WHERE btrim(coalesce("HEICode", '')) <> ''
GROUP BY lower(regexp_replace(btrim(normalize("CategoryName", NFC)), '[[:space:]]+', ' ', 'g')),
         upper(lower(regexp_replace(btrim(normalize("HEICode", NFC)), '[[:space:]]+', ' ', 'g')))
HAVING count(*) > 1;
