"use client";

import * as React from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from "recharts";

interface LatencyDistributionData {
  range: string;
  count: number;
}

interface LatencyDistributionChartProps {
  data: LatencyDistributionData[];
}

const getBarColor = (range: string) => {
  const msValue = parseInt(range.split("-")[0]) || parseInt(range.replace("+", ""));
  if (msValue < 500) return "#22c55e"; // green-500
  if (msValue < 1000) return "#84cc16"; // lime-500
  if (msValue < 2000) return "#eab308"; // yellow-500
  if (msValue < 5000) return "#f97316"; // orange-500
  return "#ef4444"; // red-500
};

export function LatencyDistributionChart({ data }: LatencyDistributionChartProps) {
  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-muted-foreground">
        No data available
      </div>
    );
  }

  return (
    <div className="h-64">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis
            dataKey="range"
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            className="text-muted-foreground"
          />
          <YAxis
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            className="text-muted-foreground"
          />
          <Tooltip
            contentStyle={{
              backgroundColor: "hsl(var(--card))",
              border: "1px solid hsl(var(--border))",
              borderRadius: "8px",
            }}
            labelStyle={{ color: "hsl(var(--foreground))" }}
            formatter={(value: number) => [value.toLocaleString(), "Requests"]}
            labelFormatter={(label) => `${label}ms`}
          />
          <Bar dataKey="count" radius={[4, 4, 0, 0]}>
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={getBarColor(entry.range)} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
