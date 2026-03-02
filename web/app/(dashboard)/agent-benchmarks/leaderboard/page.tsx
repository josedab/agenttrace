import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Agent Leaderboard | AgentTrace",
  description: "Agent performance leaderboard with ELO ratings and benchmarks",
};

export default function LeaderboardPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Agent Performance Leaderboard"
        description="Rankings with ELO-style ratings, benchmark suites, and community submissions"
      />
      <Suspense fallback={<div className="h-96 bg-muted animate-pulse rounded-lg" />}>
        <LeaderboardContent />
      </Suspense>
    </div>
  );
}

function LeaderboardContent() {
  return (
    <div className="space-y-6">
      {/* Stats overview */}
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard title="Total Agents" value="0" description="Registered agents" />
        <StatCard title="Benchmarks" value="0" description="Active benchmark suites" />
        <StatCard title="Submissions" value="0" description="Total submissions" />
        <StatCard title="Community" value="0" description="Community contributions" />
      </div>

      {/* Benchmark selector */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Select Benchmark</h3>
          <select className="rounded-md border bg-background px-3 py-2 text-sm">
            <option>All Benchmarks</option>
            <option>Code Generation</option>
            <option>Bug Fixing</option>
            <option>QA</option>
            <option>Reasoning</option>
          </select>
        </div>
      </div>

      {/* Leaderboard table */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Rankings</h3>
          <div className="flex gap-2">
            <button className="rounded-md border px-3 py-1 text-sm hover:bg-accent bg-accent">
              ELO Rating
            </button>
            <button className="rounded-md border px-3 py-1 text-sm hover:bg-accent">
              Overall Score
            </button>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b">
                <th className="text-left py-3 px-4 font-medium w-12">#</th>
                <th className="text-left py-3 px-4 font-medium">Agent</th>
                <th className="text-right py-3 px-4 font-medium">ELO Rating</th>
                <th className="text-right py-3 px-4 font-medium">Score</th>
                <th className="text-right py-3 px-4 font-medium">W/L/D</th>
                <th className="text-right py-3 px-4 font-medium">Confidence</th>
                <th className="text-center py-3 px-4 font-medium">Trend</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td colSpan={7} className="text-center py-12 text-muted-foreground">
                  <p className="text-sm">No submissions yet</p>
                  <p className="text-xs mt-1">Submit agent results to see rankings</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Community & GHA integration */}
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">Community Submissions</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Submit benchmark results via the API or GitHub Actions integration.
          </p>
          <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto">
{`POST /api/v1/benchmarks/community/submit
{
  "benchmarkId": "...",
  "agentName": "my-agent",
  "agentVersion": "1.0.0",
  "scores": { "accuracy": 0.95 }
}`}
          </pre>
        </div>

        <div className="rounded-lg border bg-card p-6">
          <h3 className="text-lg font-semibold mb-4">GitHub Actions Integration</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Automate benchmark runs with CI/CD pipelines.
          </p>
          <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto">
{`- name: Run Benchmark
  run: agenttrace benchmark run
  env:
    AGENTTRACE_API_KEY: \${{ secrets.AGENTTRACE_API_KEY }}`}
          </pre>
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, description }: { title: string; value: string; description: string }) {
  return (
    <div className="rounded-lg border bg-card p-4">
      <p className="text-sm font-medium text-muted-foreground">{title}</p>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
  );
}
