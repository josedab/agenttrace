"use client";

import * as React from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  FileCode,
  Shield,
  AlertTriangle,
  CheckCircle,
  TrendingUp,
  BarChart3,
} from "lucide-react";
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

// --- Mock Data ---

const scoreTrendData = [
  { week: "W1", score: 75 },
  { week: "W2", score: 76 },
  { week: "W3", score: 78 },
  { week: "W4", score: 79 },
  { week: "W5", score: 81 },
  { week: "W6", score: 82 },
  { week: "W7", score: 84 },
  { week: "W8", score: 85 },
];

const gradeDistribution = [
  { grade: "A", count: 12 },
  { grade: "B", count: 24 },
  { grade: "C", count: 8 },
  { grade: "D", count: 3 },
  { grade: "F", count: 1 },
];

type Severity = "blocker" | "critical" | "major" | "minor" | "info";
type Analyzer = "eslint" | "semgrep" | "sonarqube";

interface Finding {
  id: number;
  severity: Severity;
  rule: string;
  file: string;
  line: number;
  analyzer: Analyzer;
  message: string;
}

const findings: Finding[] = [
  { id: 1, severity: "blocker", rule: "no-eval", file: "src/utils/exec.ts", line: 42, analyzer: "eslint", message: "eval() usage detected — potential code injection risk" },
  { id: 2, severity: "critical", rule: "sql-injection", file: "src/db/queries.ts", line: 118, analyzer: "semgrep", message: "Unsanitized user input in SQL query" },
  { id: 3, severity: "major", rule: "cognitive-complexity", file: "src/services/agent.ts", line: 56, analyzer: "sonarqube", message: "Function cognitive complexity of 23 exceeds threshold of 15" },
  { id: 4, severity: "major", rule: "no-unused-vars", file: "src/components/Dashboard.tsx", line: 12, analyzer: "eslint", message: "'processData' is defined but never used" },
  { id: 5, severity: "minor", rule: "react-hooks/exhaustive-deps", file: "src/hooks/useAgent.ts", line: 31, analyzer: "eslint", message: "Missing dependency 'agentId' in useEffect dependency array" },
  { id: 6, severity: "critical", rule: "hardcoded-secret", file: "src/config/auth.ts", line: 8, analyzer: "semgrep", message: "Hardcoded API key detected in source code" },
  { id: 7, severity: "minor", rule: "no-console", file: "src/lib/logger.ts", line: 22, analyzer: "eslint", message: "Unexpected console.warn statement" },
  { id: 8, severity: "info", rule: "code-coverage", file: "src/services/billing.ts", line: 1, analyzer: "sonarqube", message: "Line coverage is 42%, below threshold of 80%" },
  { id: 9, severity: "major", rule: "insecure-redirect", file: "src/api/callback.ts", line: 67, analyzer: "semgrep", message: "Open redirect via unvalidated URL parameter" },
  { id: 10, severity: "info", rule: "duplicate-code", file: "src/utils/format.ts", line: 15, analyzer: "sonarqube", message: "Duplicated block of 18 lines detected across 2 files" },
];

const analyzerBreakdown = [
  { name: "ESLint", score: 88, findings: 4, color: "bg-purple-500" },
  { name: "Semgrep", score: 72, findings: 3, color: "bg-green-500" },
  { name: "SonarQube", score: 81, findings: 3, color: "bg-blue-500" },
];

const totalScans = 48;
const avgScore = 82;
const totalFindings = findings.length;
const passRate = 83;

function getGradeBadge(score: number) {
  if (score >= 90) return <Badge className="bg-green-500 text-white">A</Badge>;
  if (score >= 80) return <Badge className="bg-blue-500 text-white">B</Badge>;
  if (score >= 70) return <Badge className="bg-yellow-500 text-white">C</Badge>;
  if (score >= 60) return <Badge className="bg-orange-500 text-white">D</Badge>;
  return <Badge className="bg-red-500 text-white">F</Badge>;
}

const severityConfig: Record<Severity, string> = {
  blocker: "bg-red-600 text-white",
  critical: "bg-red-500 text-white",
  major: "bg-orange-500 text-white",
  minor: "bg-yellow-500 text-white",
  info: "bg-blue-500 text-white",
};

const analyzerConfig: Record<Analyzer, string> = {
  eslint: "bg-purple-500 text-white",
  semgrep: "bg-green-500 text-white",
  sonarqube: "bg-blue-500 text-white",
};

export function CodeQualityDashboard() {
  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Scans</CardTitle>
            <FileCode className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalScans}</div>
            <p className="text-xs text-muted-foreground">Across all repositories</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Score</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <span className="text-2xl font-bold">{avgScore}</span>
              {getGradeBadge(avgScore)}
            </div>
            <p className="text-xs text-muted-foreground">+3 from last week</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Findings</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalFindings}</div>
            <p className="text-xs text-muted-foreground">Across all analyzers</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pass Rate</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{passRate}%</div>
            <p className="text-xs text-muted-foreground">Scans meeting quality gate</p>
          </CardContent>
        </Card>
      </div>

      {/* Charts Row */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Score Trend */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <TrendingUp className="h-4 w-4" />
              Score Trend
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={scoreTrendData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis dataKey="week" className="text-xs" />
                <YAxis domain={[60, 100]} className="text-xs" />
                <Tooltip />
                <Line
                  type="monotone"
                  dataKey="score"
                  stroke="hsl(var(--primary))"
                  strokeWidth={2}
                  dot={{ r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* Grade Distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <BarChart3 className="h-4 w-4" />
              Grade Distribution
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={gradeDistribution}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis dataKey="grade" className="text-xs" />
                <YAxis className="text-xs" />
                <Tooltip />
                <Bar dataKey="count" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Analyzer Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Shield className="h-4 w-4" />
            Analyzer Breakdown
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {analyzerBreakdown.map((analyzer) => (
              <div key={analyzer.name} className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Badge className={analyzerConfig[analyzer.name.toLowerCase() as Analyzer]}>
                      {analyzer.name}
                    </Badge>
                    <span className="text-sm text-muted-foreground">
                      {analyzer.findings} findings
                    </span>
                  </div>
                  <span className="text-sm font-medium">{analyzer.score}/100</span>
                </div>
                <div className="h-2 w-full rounded-full bg-muted">
                  <div
                    className={`h-2 rounded-full ${analyzer.color}`}
                    style={{ width: `${analyzer.score}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Findings Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <AlertTriangle className="h-4 w-4" />
            Findings
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-muted-foreground">
                  <th className="pb-2 pr-4 font-medium">Severity</th>
                  <th className="pb-2 pr-4 font-medium">Rule</th>
                  <th className="pb-2 pr-4 font-medium">File</th>
                  <th className="pb-2 pr-4 font-medium">Line</th>
                  <th className="pb-2 pr-4 font-medium">Analyzer</th>
                  <th className="pb-2 font-medium">Message</th>
                </tr>
              </thead>
              <tbody>
                {findings.map((f) => (
                  <tr key={f.id} className="border-b last:border-0">
                    <td className="py-2 pr-4">
                      <Badge className={severityConfig[f.severity]}>
                        {f.severity}
                      </Badge>
                    </td>
                    <td className="py-2 pr-4 font-mono text-xs">{f.rule}</td>
                    <td className="py-2 pr-4 font-mono text-xs">{f.file}</td>
                    <td className="py-2 pr-4 text-center">{f.line}</td>
                    <td className="py-2 pr-4">
                      <Badge className={analyzerConfig[f.analyzer]}>
                        {f.analyzer}
                      </Badge>
                    </td>
                    <td className="py-2 text-muted-foreground">{f.message}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
