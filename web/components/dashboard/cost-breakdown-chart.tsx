"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

interface CostBreakdownData {
  model: string;
  cost: number;
  tokens: number;
}

interface CostBreakdownChartProps {
  data: CostBreakdownData[];
}

const COLORS = [
  "hsl(var(--chart-1))",
  "hsl(var(--chart-2))",
  "hsl(var(--chart-3))",
  "hsl(var(--chart-4))",
  "hsl(var(--chart-5))",
];

export function CostBreakdownChart({ data }: CostBreakdownChartProps) {
  // Sort by cost and take top 5
  const sortedData = [...data]
    .sort((a, b) => b.cost - a.cost)
    .slice(0, 5);

  const totalCost = data.reduce((sum, item) => sum + item.cost, 0);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Cost by Model</CardTitle>
        <CardDescription>
          Top models by cost (Total: ${totalCost.toFixed(2)})
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={sortedData} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" horizontal={false} />
              <XAxis
                type="number"
                tickLine={false}
                axisLine={false}
                className="text-xs text-muted-foreground"
                tickFormatter={(value) => `$${value.toFixed(2)}`}
              />
              <YAxis
                type="category"
                dataKey="model"
                tickLine={false}
                axisLine={false}
                className="text-xs text-muted-foreground"
                width={120}
                tickFormatter={(value) => {
                  // Truncate long model names
                  if (value.length > 15) {
                    return value.slice(0, 15) + "...";
                  }
                  return value;
                }}
              />
              <Tooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length) return null;
                  const data = payload[0].payload;
                  return (
                    <div className="rounded-lg border bg-background p-3 shadow-md">
                      <p className="text-sm font-medium mb-2">{data.model}</p>
                      <div className="space-y-1 text-sm">
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Cost:</span>
                          <span className="font-medium">
                            ${data.cost.toFixed(4)}
                          </span>
                        </div>
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Tokens:</span>
                          <span className="font-medium">
                            {data.tokens.toLocaleString()}
                          </span>
                        </div>
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Share:</span>
                          <span className="font-medium">
                            {((data.cost / totalCost) * 100).toFixed(1)}%
                          </span>
                        </div>
                      </div>
                    </div>
                  );
                }}
              />
              <Bar dataKey="cost" radius={[0, 4, 4, 0]}>
                {sortedData.map((_, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
