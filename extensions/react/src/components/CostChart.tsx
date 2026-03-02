import React, { useMemo } from "react";

export interface CostChartProps {
  /** Cost data points */
  data: CostDataPoint[];
  /** Chart height */
  height?: number;
  /** Chart width */
  width?: number;
  /** Show labels */
  showLabels?: boolean;
  /** Theme */
  theme?: "light" | "dark";
  /** Currency symbol */
  currency?: string;
  /** Custom class name */
  className?: string;
}

export interface CostDataPoint {
  label: string;
  value: number;
  color?: string;
  metadata?: Record<string, unknown>;
}

export function CostChart({
  data,
  height = 300,
  width = 600,
  showLabels = true,
  theme = "light",
  currency = "$",
  className = "",
}: CostChartProps) {
  const isDark = theme === "dark";
  const maxValue = useMemo(() => Math.max(...data.map((d) => d.value), 0.01), [data]);
  const totalCost = useMemo(() => data.reduce((sum, d) => sum + d.value, 0), [data]);

  const barWidth = Math.max(20, Math.min(60, (width - 80) / data.length - 8));
  const chartHeight = height - 80;

  const colors = [
    "#3b82f6", "#10b981", "#f59e0b", "#ef4444",
    "#8b5cf6", "#ec4899", "#06b6d4", "#84cc16",
  ];

  const bg = isDark ? "#1a1a2e" : "#ffffff";
  const border = isDark ? "#2d2d44" : "#e5e7eb";
  const text = isDark ? "#e5e7eb" : "#1f2937";
  const muted = isDark ? "#9ca3af" : "#6b7280";

  if (data.length === 0) {
    return (
      <div
        style={{
          width,
          height,
          backgroundColor: bg,
          border: `1px solid ${border}`,
          borderRadius: "8px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: muted,
          fontFamily: 'sans-serif',
        }}
        className={className}
      >
        No cost data available
      </div>
    );
  }

  return (
    <div
      style={{
        width,
        backgroundColor: bg,
        border: `1px solid ${border}`,
        borderRadius: "8px",
        padding: "16px",
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        color: text,
      }}
      className={className}
    >
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "12px" }}>
        <span style={{ fontWeight: 600, fontSize: "14px" }}>Cost Breakdown</span>
        <span style={{ fontFamily: "monospace", fontSize: "14px" }}>
          Total: {currency}{totalCost.toFixed(4)}
        </span>
      </div>

      {/* Chart */}
      <svg width={width - 32} height={chartHeight} role="img" aria-label="Cost chart">
        {data.map((d, i) => {
          const barHeight = (d.value / maxValue) * (chartHeight - 30);
          const x = i * (barWidth + 8) + 40;
          const y = chartHeight - barHeight - 20;
          const color = d.color || colors[i % colors.length];

          return (
            <g key={i}>
              <rect
                x={x}
                y={y}
                width={barWidth}
                height={barHeight}
                fill={color}
                rx={4}
                opacity={0.85}
              />
              {/* Value label */}
              <text
                x={x + barWidth / 2}
                y={y - 4}
                textAnchor="middle"
                fontSize="11"
                fill={muted}
                fontFamily="monospace"
              >
                {currency}{d.value.toFixed(4)}
              </text>
              {/* X-axis label */}
              {showLabels && (
                <text
                  x={x + barWidth / 2}
                  y={chartHeight - 4}
                  textAnchor="middle"
                  fontSize="10"
                  fill={muted}
                >
                  {d.label.length > 10 ? d.label.slice(0, 10) + "…" : d.label}
                </text>
              )}
            </g>
          );
        })}
      </svg>
    </div>
  );
}
