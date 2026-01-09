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

interface ScoreDistributionData {
  range: string;
  count: number;
}

interface ScoreDistributionChartProps {
  data: ScoreDistributionData[];
}

const getBarColor = (range: string) => {
  const rangeValue = parseFloat(range.split("-")[0]);
  if (rangeValue >= 0.8) return "#22c55e"; // green-500
  if (rangeValue >= 0.6) return "#84cc16"; // lime-500
  if (rangeValue >= 0.4) return "#eab308"; // yellow-500
  if (rangeValue >= 0.2) return "#f97316"; // orange-500
  return "#ef4444"; // red-500
};

export function ScoreDistributionChart({ data }: ScoreDistributionChartProps) {
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

// Alternative simple distribution without recharts
export function SimpleScoreDistribution({ scores }: { scores: number[] }) {
  if (!scores || scores.length === 0) {
    return (
      <div className="flex items-center justify-center h-32 text-muted-foreground text-sm">
        No scores available
      </div>
    );
  }

  const bins = [0, 0, 0, 0, 0]; // 0-0.2, 0.2-0.4, 0.4-0.6, 0.6-0.8, 0.8-1.0
  scores.forEach((score) => {
    const index = Math.min(Math.floor(score * 5), 4);
    bins[index]++;
  });

  const maxCount = Math.max(...bins);
  const binLabels = ["0-0.2", "0.2-0.4", "0.4-0.6", "0.6-0.8", "0.8-1.0"];
  const binColors = [
    "bg-red-500",
    "bg-orange-500",
    "bg-yellow-500",
    "bg-lime-500",
    "bg-green-500",
  ];

  return (
    <div className="flex items-end gap-2 h-32">
      {bins.map((count, index) => (
        <div key={index} className="flex-1 flex flex-col items-center gap-1">
          <div
            className={`w-full rounded-t ${binColors[index]}`}
            style={{
              height: maxCount > 0 ? `${(count / maxCount) * 100}%` : 0,
              minHeight: count > 0 ? "4px" : 0,
            }}
          />
          <span className="text-xs text-muted-foreground">{binLabels[index]}</span>
          <span className="text-xs font-medium">{count}</span>
        </div>
      ))}
    </div>
  );
}
