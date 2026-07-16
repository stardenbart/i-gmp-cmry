import { cn } from "@/lib/utils";
import { ArrowUpRight, ArrowDownRight, Loader2 } from "lucide-react";
import { Card } from "../ui/card";

interface StatCardProps {
  title: string;
  value: number | string;
  icon: React.ElementType;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  color: "primary" | "orange" | "green" | "purple" | "blue";
  isLoading?: boolean;
}

const colorMap = {
  primary: "text-primary bg-primary/10",
  orange: "text-orange-500 bg-orange-500/10",
  green: "text-green-500 bg-green-500/10",
  purple: "text-purple-500 bg-purple-500/10",
  blue: "text-blue-500 bg-blue-500/10",
};

export function StatCard({ title, value, icon: Icon, trend, color, isLoading }: StatCardProps) {
  return (
    <Card className="p-4 sm:p-6 bg-card/60 backdrop-blur-md hover:bg-card/80 transition-colors">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <div className={cn("p-2 rounded-lg", colorMap[color])}>
            <Icon className="h-5 w-5" />
          </div>
          {trend && (
            <div
              className={cn(
                "flex items-center gap-1 text-xs font-medium px-2 py-1 rounded-full",
                trend.isPositive
                  ? "text-green-500 bg-green-500/10"
                  : "text-red-500 bg-red-500/10"
              )}
            >
              {trend.isPositive ? (
                <ArrowUpRight className="h-3 w-3" />
              ) : (
                <ArrowDownRight className="h-3 w-3" />
              )}
              {Math.abs(trend.value)}%
            </div>
          )}
        </div>
        {isLoading ? (
          <div className="flex items-center gap-2">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            <span className="text-2xl font-bold text-muted-foreground">—</span>
          </div>
        ) : (
          <span className="text-2xl font-bold">{value.toLocaleString("id-ID")}</span>
        )}
        <span className="text-xs text-muted-foreground">{title}</span>
      </div>
    </Card>
  );
}