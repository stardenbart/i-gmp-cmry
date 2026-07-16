import { StatCard } from "../ui/state-card"
import {
  ClipboardCheck,
  AlertTriangle,
  Users,
  History,
} from "lucide-react";


export interface DashboardStats {
  total_inspections: number;
  open_issues: number;
  closed_issues: number;
  total_pics: number;
  inspections_this_month: number;
  issues_resolved_this_month: number;
  issues_by_month: Array<{ month: string; count: number }>;
  issues_trend: Array<{ date: string; open: number; closed: number }>;
  recent_inspections: Array<{
    id: string;
    name: string;
    status: string;
    created_at: string;
    area_name?: string;
  }>;
  top_issues_by_area: Array<{ area_name: string; count: number }>;
}

export const StatsCards = ({ stats, isLoading }: { stats: DashboardStats | undefined, isLoading: boolean }) => {
    return (
        <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
                <StatCard
                  title="Total Inspeksi"
                  value={stats?.total_inspections ?? 0}
                  icon={ClipboardCheck}
                  trend={
                    stats?.inspections_this_month
                      ? {
                          value: Math.round((stats.inspections_this_month / (stats.total_inspections || 1)) * 100),
                          isPositive: true,
                        }
                      : undefined
                  }
                  color="primary"
                  isLoading={isLoading}
                />
                <StatCard
                  title="Open Issues"
                  value={stats?.open_issues ?? 0}
                  icon={AlertTriangle}
                  color="orange"
                  isLoading={isLoading}
                />
                <StatCard
                  title="Closed Issues"
                  value={stats?.closed_issues ?? 0}
                  icon={History}
                  trend={
                    stats?.issues_resolved_this_month
                      ? {
                          value: Math.round((stats.issues_resolved_this_month / (stats.closed_issues || 1)) * 100),
                          isPositive: true,
                        }
                      : undefined
                  }
                  color="green"
                  isLoading={isLoading}
                />
                <StatCard
                  title="Total PIC"
                  value={stats?.total_pics ?? 0}
                  icon={Users}
                  color="purple"
                  isLoading={isLoading}
                />
              </div>
    )
}