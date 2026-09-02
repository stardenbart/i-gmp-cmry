"use client";

import {
  ResponsiveContainer,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  Bar,
  BarChart as RechartsBarChart,
  Line,
  LineChart as RechartsLineChart,
  Pie,
  PieChart as RechartsPieChart,
  Cell,
} from "recharts";
import { AreaChart } from "@/components/ui/area-chart";
import type { VizType } from "@/components/dashboard/types";

export interface VisualizationValueKey {
  key: string;
  name: string;
  color: string;
}

interface VisualizationSwitchProps {
  /** Must not be "list" — callers keep their own bespoke markup for that. */
  vizType: Exclude<VizType, "list">;
  data: Array<Record<string, string | number>>;
  categoryKey: string;
  valueKeys: VisualizationValueKey[];
  height?: number;
}

const AXIS_STYLE = { stroke: "#71717a", fontSize: 12, tickLine: false, axisLine: false } as const;
const TOOLTIP_STYLE = { contentStyle: { backgroundColor: "#09090b", borderColor: "#27272a", borderRadius: "8px" }, itemStyle: { color: "#fafafa" } };

function SimpleTable({ data, categoryKey, valueKeys }: Pick<VisualizationSwitchProps, "data" | "categoryKey" | "valueKeys">) {
  return (
    <div className="h-full overflow-auto rounded-lg border border-border">
      <table className="w-full text-left text-xs">
        <thead className="sticky top-0 bg-muted/60">
          <tr>
            <th className="px-3 py-2 font-semibold text-muted-foreground">{categoryKey}</th>
            {valueKeys.map((v) => (
              <th key={v.key} className="px-3 py-2 font-semibold text-muted-foreground">{v.name}</th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {data.map((row, idx) => (
            <tr key={idx} className="hover:bg-muted/30">
              <td className="px-3 py-2 font-medium text-foreground">{String(row[categoryKey] ?? "")}</td>
              {valueKeys.map((v) => (
                <td key={v.key} className="px-3 py-2 text-foreground">{String(row[v.key] ?? "")}</td>
              ))}
            </tr>
          ))}
          {data.length === 0 && (
            <tr>
              <td colSpan={valueKeys.length + 1} className="px-3 py-6 text-center italic text-muted-foreground">Belum ada data</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

/** Renders the same underlying data as a bar/line/area chart, table, or pie
 * — the Power BI-style "change visualization" picker in the widget header
 * (see DashboardWidgetFrame) swaps which of these gets used. */
export function VisualizationSwitch({ vizType, data, categoryKey, valueKeys, height = 260 }: VisualizationSwitchProps) {
  if (vizType === "table") {
    return <SimpleTable data={data} categoryKey={categoryKey} valueKeys={valueKeys} />;
  }

  if (vizType === "area") {
    return (
      <AreaChart
        data={data}
        xAxisKey={categoryKey}
        height={height}
        series={valueKeys.map((v) => ({ dataKey: v.key, name: v.name, color: v.color }))}
      />
    );
  }

  if (vizType === "pie") {
    // A pie only makes sense for one series — the first valueKey wins.
    const [primary] = valueKeys;
    if (!primary) return null;
    return (
      <ResponsiveContainer width="100%" height={height}>
        <RechartsPieChart>
          <Tooltip {...TOOLTIP_STYLE} />
          <Legend wrapperStyle={{ fontSize: 12 }} />
          <Pie data={data} dataKey={primary.key} nameKey={categoryKey} outerRadius="80%" label>
            {data.map((_, idx) => (
              <Cell key={idx} fill={PIE_COLORS[idx % PIE_COLORS.length]} />
            ))}
          </Pie>
        </RechartsPieChart>
      </ResponsiveContainer>
    );
  }

  if (vizType === "line") {
    return (
      <ResponsiveContainer width="100%" height={height}>
        <RechartsLineChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#27272a" vertical={false} />
          <XAxis dataKey={categoryKey} {...AXIS_STYLE} interval="preserveStartEnd" minTickGap={24} />
          <YAxis {...AXIS_STYLE} />
          <Tooltip {...TOOLTIP_STYLE} />
          <Legend wrapperStyle={{ fontSize: 12 }} />
          {valueKeys.map((v) => (
            <Line key={v.key} type="monotone" dataKey={v.key} name={v.name} stroke={v.color} strokeWidth={2} dot={false} />
          ))}
        </RechartsLineChart>
      </ResponsiveContainer>
    );
  }

  // "bar" (and any unrecognized value) — the safest, most generic fallback.
  return (
    <ResponsiveContainer width="100%" height={height}>
      <RechartsBarChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#27272a" vertical={false} />
        <XAxis dataKey={categoryKey} {...AXIS_STYLE} interval="preserveStartEnd" minTickGap={24} />
        <YAxis {...AXIS_STYLE} />
        <Tooltip {...TOOLTIP_STYLE} />
        <Legend wrapperStyle={{ fontSize: 12 }} />
        {valueKeys.map((v) => (
          <Bar key={v.key} dataKey={v.key} name={v.name} fill={v.color} radius={[4, 4, 0, 0]} />
        ))}
      </RechartsBarChart>
    </ResponsiveContainer>
  );
}

const PIE_COLORS = ["#2563eb", "#10b981", "#a855f7", "#f97316", "#ef4444", "#eab308", "#14b8a6", "#f43f5e"];
