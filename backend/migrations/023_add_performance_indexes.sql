-- Migration 023: Add performance indexes for high-frequency queries
CREATE INDEX IF NOT EXISTS idx_users_userid ON "Users" ("UserID");
CREATE INDEX IF NOT EXISTS idx_users_username ON "Users" ("Username");
CREATE INDEX IF NOT EXISTS idx_activitylog_createdat ON "Activity_Log" ("ActivityCreatedAt" DESC);
CREATE INDEX IF NOT EXISTS idx_activitylog_userid ON "Activity_Log" ("UserID");
CREATE INDEX IF NOT EXISTS idx_hei_category_code ON "HEI_Master" ("CategoryName", "HEICode");
CREATE INDEX IF NOT EXISTS idx_pic_mapping_userid ON "PIC_Mapping" ("UserID");
CREATE INDEX IF NOT EXISTS idx_kawasan_aspek_kawasanid ON "kawasan_aspek" ("KawasanID");
