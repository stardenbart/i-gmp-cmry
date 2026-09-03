-- User_Dashboard_Layout previously had UserID alone as its primary key — one
-- shared layout row per user across every dashboard. This adds DashboardKey
-- so a user can have several independently-saved layouts (their role
-- dashboard, "main"; the new KPI/analytics dashboard, "kpi"; etc.) without
-- one overwriting another.
--
-- Default 'main' makes every existing row automatically satisfy the new
-- composite primary key uniquely (old PK was UserID alone, so UserID was
-- already unique) — no data loss, no downtime, and every pre-existing role
-- dashboard keeps behaving exactly as before (callers that don't pass
-- ?dashboard_key= default to "main" server-side, see dashboard_handler.go).
ALTER TABLE "User_Dashboard_Layout" ADD COLUMN IF NOT EXISTS "DashboardKey" VARCHAR(50) NOT NULL DEFAULT 'main';
ALTER TABLE "User_Dashboard_Layout" DROP CONSTRAINT IF EXISTS "User_Dashboard_Layout_pkey";
ALTER TABLE "User_Dashboard_Layout" ADD CONSTRAINT "User_Dashboard_Layout_pkey" PRIMARY KEY ("UserID", "DashboardKey");
