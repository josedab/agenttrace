"use client";

import * as React from "react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { format } from "date-fns";

interface TokenUsageData {
  date: string;
  inputTokens: number;
  outputTokens: number;
}

interface TokenUsageChartProps {
  data: TokenUsageData[];
}

export function TokenUsageChart({ data }: TokenUsageChartProps) {
  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-muted-foreground">
        No data available
      </div>
    );
  }

  const formatTokens = (value: number) => {
    if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
    if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`;
    return value.toString();
  };

  return (
    <div className="h-64">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="colorInput" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorOutput" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis
            dataKey="date"
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            tickFormatter={(value) => format(new Date(value), "MMM d")}
            className="text-muted-foreground"
          />
          <YAxis
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            tickFormatter={formatTokens}
            className="text-muted-foreground"
          />
          <Tooltip
            contentStyle={{
              backgroundColor: "hsl(var(--card))",
              border: "1px solid hsl(var(--border))",
              borderRadius: "8px",
            }}
            labelStyle={{ color: "hsl(var(--foreground))" }}
            formatter={(value: number, name: string) => [
              value.toLocaleString(),
              name === "inputTokens" ? "Input Tokens" : "Output Tokens",
            ]}
            labelFormatter={(label) => format(new Date(label), "MMM d, yyyy")}
          />
          <Legend
            formatter={(value) =>
              value === "inputTokens" ? "Input Tokens" : "Output Tokens"
            }
          />
          <Area
            type="monotone"
            dataKey="inputTokens"
            stackId="1"
            stroke="#3b82f6"
            fillOpacity={1}
            fill="url(#colorInput)"
          />
          <Area
            type="monotone"
            dataKey="outputTokens"
            stackId="1"
            stroke="#8b5cf6"
            fillOpacity={1}
            fill="url(#colorOutput)"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
