"use client";

import * as React from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  TrendingUp,
  Calculator,
  Wallet,
  Plus,
  Trash2,
  Play,
  DollarSign,
  ArrowDown,
  ArrowUp,
} from "lucide-react";

interface ForecastPoint {
  date: string;
  predictedCost: number;
  lowerBound: number;
  upperBound: number;
}

interface SimulationChange {
  fromModel: string;
  toModel: string;
  trafficPercent: number;
}

interface SimulationResult {
  baselineCost: number;
  projectedCost: number;
  savings: number;
  qualityImpact: number;
}

export function CostForecastDashboard() {
  const [period, setPeriod] = React.useState("daily");
  const [changes, setChanges] = React.useState<SimulationChange[]>([]);
  const [budgetForm, setBudgetForm] = React.useState({
    monthlyBudget: "",
    alertThreshold80: true,
    alertThreshold90: true,
    alertThreshold100: true,
    modelAllocations: "",
  });

  const { data: forecast, isLoading: forecastLoading } = useQuery<ForecastPoint[]>({
    queryKey: ["cost-forecast", period],
    queryFn: () => api.costForecast.getForecast({ period }),
  });

  const simulateMutation = useMutation<SimulationResult, Error, SimulationChange[]>({
    mutationFn: (simChanges) => api.costForecast.simulate({ changes: simChanges }),
  });

  const budgetMutation = useMutation({
    mutationFn: (data: typeof budgetForm) => api.costForecast.createBudget(data),
  });

  const addChange = () => {
    setChanges([...changes, { fromModel: "", toModel: "", trafficPercent: 100 }]);
  };

  const removeChange = (index: number) => {
    setChanges(changes.filter((_, i) => i !== index));
  };

  const updateChange = (index: number, field: keyof SimulationChange, value: string | number) => {
    const updated = [...changes];
    updated[index] = { ...updated[index], [field]: value };
    setChanges(updated);
  };

  return (
    <Tabs defaultValue="forecast" className="space-y-6">
      <TabsList>
        <TabsTrigger value="forecast" className="flex items-center gap-1">
          <TrendingUp className="h-4 w-4" />
          Forecast
        </TabsTrigger>
        <TabsTrigger value="simulator" className="flex items-center gap-1">
          <Calculator className="h-4 w-4" />
          Simulator
        </TabsTrigger>
        <TabsTrigger value="budget" className="flex items-center gap-1">
          <Wallet className="h-4 w-4" />
          Budget
        </TabsTrigger>
      </TabsList>

      {/* Forecast Tab */}
      <TabsContent value="forecast" className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium">Cost Forecast</h3>
          <Select value={period} onValueChange={setPeriod}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="daily">Daily</SelectItem>
              <SelectItem value="weekly">Weekly</SelectItem>
              <SelectItem value="monthly">Monthly</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <Card>
          <CardContent className="pt-6">
            {forecastLoading ? (
              <div className="h-64 bg-muted animate-pulse rounded-lg" />
            ) : forecast && forecast.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left py-2 font-medium">Date</th>
                      <th className="text-right py-2 font-medium">Predicted Cost</th>
                      <th className="text-right py-2 font-medium">Lower Bound</th>
                      <th className="text-right py-2 font-medium">Upper Bound</th>
                    </tr>
                  </thead>
                  <tbody>
                    {forecast.map((point) => (
                      <tr key={point.date} className="border-b last:border-0">
                        <td className="py-2 text-muted-foreground">{point.date}</td>
                        <td className="py-2 text-right font-medium">
                          ${point.predictedCost.toFixed(2)}
                        </td>
                        <td className="py-2 text-right text-muted-foreground">
                          ${point.lowerBound.toFixed(2)}
                        </td>
                        <td className="py-2 text-right text-muted-foreground">
                          ${point.upperBound.toFixed(2)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground text-sm">
                No forecast data available
              </div>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      {/* Simulator Tab */}
      <TabsContent value="simulator" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Calculator className="h-4 w-4" />
              Model Routing Changes
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {changes.map((change, index) => (
              <div key={index} className="flex items-center gap-3">
                <Input
                  placeholder="From model"
                  value={change.fromModel}
                  onChange={(e) => updateChange(index, "fromModel", e.target.value)}
                  className="flex-1"
                />
                <span className="text-muted-foreground text-sm">→</span>
                <Input
                  placeholder="To model"
                  value={change.toModel}
                  onChange={(e) => updateChange(index, "toModel", e.target.value)}
                  className="flex-1"
                />
                <Input
                  type="number"
                  placeholder="%"
                  value={change.trafficPercent}
                  onChange={(e) => updateChange(index, "trafficPercent", parseInt(e.target.value) || 0)}
                  className="w-20"
                />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => removeChange(index)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}

            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={addChange}>
                <Plus className="h-4 w-4 mr-1" />
                Add Change
              </Button>
              {changes.length > 0 && (
                <Button
                  size="sm"
                  onClick={() => simulateMutation.mutate(changes)}
                  disabled={simulateMutation.isPending}
                >
                  <Play className="h-4 w-4 mr-1" />
                  Simulate
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Simulation Results */}
        {simulateMutation.data && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Card>
              <CardContent className="pt-4">
                <div className="text-xs text-muted-foreground">Baseline Cost</div>
                <div className="text-xl font-bold flex items-center gap-1">
                  <DollarSign className="h-4 w-4" />
                  {simulateMutation.data.baselineCost.toFixed(2)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-4">
                <div className="text-xs text-muted-foreground">Projected Cost</div>
                <div className="text-xl font-bold flex items-center gap-1">
                  <DollarSign className="h-4 w-4" />
                  {simulateMutation.data.projectedCost.toFixed(2)}
                </div>
              </CardContent>
            </Card>
            <Card className="border-green-200 dark:border-green-800">
              <CardContent className="pt-4">
                <div className="text-xs text-muted-foreground">Savings</div>
                <div className="text-xl font-bold text-green-600 flex items-center gap-1">
                  <ArrowDown className="h-4 w-4" />
                  ${simulateMutation.data.savings.toFixed(2)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="pt-4">
                <div className="text-xs text-muted-foreground">Quality Impact</div>
                <div className={`text-xl font-bold flex items-center gap-1 ${
                  simulateMutation.data.qualityImpact >= 0 ? "text-green-600" : "text-red-600"
                }`}>
                  {simulateMutation.data.qualityImpact >= 0 ? (
                    <ArrowUp className="h-4 w-4" />
                  ) : (
                    <ArrowDown className="h-4 w-4" />
                  )}
                  {Math.abs(simulateMutation.data.qualityImpact).toFixed(1)}%
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </TabsContent>

      {/* Budget Tab */}
      <TabsContent value="budget" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Wallet className="h-4 w-4" />
              Create Budget Plan
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1">
                <label className="text-sm font-medium">Monthly Budget ($)</label>
                <Input
                  type="number"
                  placeholder="1000.00"
                  value={budgetForm.monthlyBudget}
                  onChange={(e) => setBudgetForm({ ...budgetForm, monthlyBudget: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium">Model Allocations</label>
                <Input
                  placeholder='e.g. gpt-4:40%, gpt-3.5:60%'
                  value={budgetForm.modelAllocations}
                  onChange={(e) => setBudgetForm({ ...budgetForm, modelAllocations: e.target.value })}
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">Alert Thresholds</label>
              <div className="flex gap-3">
                {[80, 90, 100].map((threshold) => {
                  const key = `alertThreshold${threshold}` as keyof typeof budgetForm;
                  return (
                    <Badge
                      key={threshold}
                      variant={budgetForm[key] ? "default" : "outline"}
                      className="cursor-pointer"
                      onClick={() =>
                        setBudgetForm({ ...budgetForm, [key]: !budgetForm[key] })
                      }
                    >
                      {threshold}%
                    </Badge>
                  );
                })}
              </div>
            </div>

            <div className="flex justify-end">
              <Button
                size="sm"
                onClick={() => budgetMutation.mutate(budgetForm)}
                disabled={budgetMutation.isPending || !budgetForm.monthlyBudget}
              >
                Create Budget Plan
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  );
}
