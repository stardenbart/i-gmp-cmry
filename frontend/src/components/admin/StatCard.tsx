import {
  Loader2,
  TrendingUp,
} from "lucide-react";

export interface DashboardStats {
  total_inspections_running: number;
  compliance_rate_trend: number;
  compliance_rate: number;
  total_open_issues: number;
  issue_overdue_trend: number;
  issue_overdue: number;
}

interface StatCardConfig {
  id: string;
  title: string;
  value: keyof DashboardStats;
  color: string;
  showLive?: boolean;
  showTrend?: boolean;
  trendValue?: keyof DashboardStats;
  isPercentage?: boolean;
  showIcon?: boolean;
  iconBg?: string;
  iconColor?: string;
  iconComponent?: React.ReactNode;
  badgeText?: string;
  badgeColor?: string;
}

const StatCard = ({
  title,
  value,
  color,
  showLive,
  showTrend,
  trendValue,
  isPercentage,
  badgeText,
  badgeColor,
  stats,
  isLoading
}: {
  title: string;
  value: keyof DashboardStats;
  color: string;
  showLive?: boolean;
  showTrend?: boolean;
  trendValue?: keyof DashboardStats;
  isPercentage?: boolean;
  badgeText?: string;
  badgeColor?: string;
  stats: DashboardStats | undefined;
  isLoading: boolean;
}) => {
  const displayValue = stats?.[value] ?? 0;
  const trendDisplay = trendValue ? stats?.[trendValue] : undefined;

  return (
    <div className="bg-card rounded-xl p-4 border border-border shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group">
      <div className={`absolute top-0 left-0 w-full h-1 ${color}`}></div>
      
      <div className="flex justify-between items-start mb-3">
        {showLive && (
          <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">Aktif</span>
        )}
        {showTrend && trendDisplay !== undefined && (
          <span className={`text-xs font-semibold ${badgeColor || 'text-green-600'} flex items-center gap-1`}>
            <TrendingUp className="h-3 w-3" /> {isPercentage ? `${Number(trendDisplay).toFixed(1)}%` : trendDisplay}
          </span>
        )}
        {badgeText && !showTrend && (
          <span className={`text-[11px] font-semibold ${badgeColor || 'text-muted-foreground'}`}>
            {badgeText}
          </span>
        )}
      </div>

      <h3 className="text-xs font-semibold text-muted-foreground mb-1 uppercase tracking-wider">
        {title}
      </h3>
      
      <p className="text-2xl font-bold text-foreground">
        {isLoading ? (
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        ) : (
          isPercentage && value === 'compliance_rate'
            ? `${Number(displayValue).toFixed(1)}%`
            : displayValue
        )}
      </p>
    </div>
  );
};

export const StatsCards = ({ 
  stats, 
  isLoading 
}: { 
  stats: DashboardStats | undefined; 
  isLoading: boolean;
}) => {
  const cardConfigs: StatCardConfig[] = [
    {
      id: 'running',
      title: 'Inspeksi Berlangsung',
      value: 'total_inspections_running',
      color: 'bg-primary',
      showLive: true
    },
    {
      id: 'compliance',
      title: 'Tingkat Kepatuhan',
      value: 'compliance_rate',
      color: 'bg-green-600',
      showTrend: true,
      trendValue: 'compliance_rate_trend',
      isPercentage: true,
      badgeColor: 'text-green-600'
    },
    {
      id: 'open-issues',
      title: 'Temuan Terbuka',
      value: 'total_open_issues',
      color: 'bg-amber-500',
      badgeText: 'Perlu Tindakan',
      badgeColor: 'text-amber-600'
    },
    {
      id: 'overdue',
      title: 'Temuan Jatuh Tempo',
      value: 'issue_overdue',
      color: 'bg-red-500',
      showTrend: true,
      trendValue: 'issue_overdue_trend',
      badgeColor: 'text-red-500',
    }
  ];

  return (
    <>
      {cardConfigs.map((config) => (
        <StatCard
          key={config.id}
          title={config.title}
          value={config.value}
          color={config.color}
          showLive={config.showLive}
          showTrend={config.showTrend}
          trendValue={config.trendValue}
          isPercentage={config.isPercentage}
          badgeText={config.badgeText}
          badgeColor={config.badgeColor}
          stats={stats}
          isLoading={isLoading}
        />
      ))}
    </>
  );
};