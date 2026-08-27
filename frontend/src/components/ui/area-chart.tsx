"use client";

import { ResponsiveContainer, CartesianGrid, XAxis, YAxis, Tooltip, Area, AreaChart as RechartsAreaChart } from "recharts"

export interface AreaSeries {
  dataKey: string;
  name?: string;
  color: string;
  unit?: string;
  yAxisId?: "left" | "right";
}

export interface AreaChartProps {
  data: Array<Record<string, string | number>>;
  xAxisKey: string;
  series: AreaSeries[];
  height?: number;
  unit?: string;
}

export function AreaChart({ data, xAxisKey, series, height = 300, unit = "" }: AreaChartProps) {
  // Map series key to custom unit
  const seriesUnitMap: Record<string, string> = {};
  series.forEach((s) => {
    seriesUnitMap[s.dataKey] = s.unit !== undefined ? s.unit : unit;
  });
  const hasRightAxis = series.some((item) => item.yAxisId === "right");
  const leftAxisUnit = series.find((item) => (item.yAxisId || "left") === "left")?.unit ?? unit;
  const rightAxisUnit = series.find((item) => item.yAxisId === "right")?.unit ?? "";

  return (
    <ResponsiveContainer width="100%" height={height}>
      <RechartsAreaChart data={data} margin={{ top: 10, right: hasRightAxis ? 0 : 10, left: -20, bottom: 0 }}>
        <defs>
          {series.map((s, idx) => (
            <linearGradient key={idx} id={`color-${s.dataKey}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={s.color} stopOpacity={0.3} />
              <stop offset="95%" stopColor={s.color} stopOpacity={0} />
            </linearGradient>
          ))}
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="#27272a" vertical={false} />
        <XAxis
          dataKey={xAxisKey}
          stroke="#71717a"
          fontSize={12}
          tickLine={false}
          axisLine={false}
          interval="preserveStartEnd"
          minTickGap={24}
        />
        <YAxis
          yAxisId="left"
          stroke="#71717a"
          fontSize={12}
          tickLine={false}
          axisLine={false}
          tickFormatter={(value) => `${value}${leftAxisUnit}`}
        />
        {hasRightAxis && (
          <YAxis
            yAxisId="right"
            orientation="right"
            stroke="#71717a"
            fontSize={12}
            tickLine={false}
            axisLine={false}
            allowDecimals={false}
            tickFormatter={(value) => `${value}${rightAxisUnit}`}
          />
        )}
        <Tooltip
          contentStyle={{
            backgroundColor: "#09090b",
            borderColor: "#27272a",
            borderRadius: "8px",
          }}
          itemStyle={{ color: "#fafafa" }}
          formatter={(value, name, item) => {
            const dataKey = typeof item?.dataKey === "string" ? item.dataKey : "";
            const seriesUnit = seriesUnitMap[dataKey] !== undefined ? seriesUnitMap[dataKey] : unit;
            return [`${value}${seriesUnit}`, name || "Nilai"];
          }}
        />
        {series.map((s, idx) => (
          <Area
            key={idx}
            type="monotone"
            dataKey={s.dataKey}
            yAxisId={s.yAxisId || "left"}
            stroke={s.color}
            fillOpacity={1}
            fill={`url(#color-${s.dataKey})`}
            name={s.name || s.dataKey}
          />
        ))}
      </RechartsAreaChart>
    </ResponsiveContainer>
  );
}
