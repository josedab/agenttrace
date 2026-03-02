import { Suspense } from "react";
import { PageHeader } from "@/components/layout/page-header";

export const metadata = {
  title: "Streaming Trace Debugger | AgentTrace",
  description: "Real-time step-through debugger for live agent sessions",
};

export default function DebuggerPage() {
  return (
    <div className="space-y-4">
      <PageHeader
        title="Streaming Trace Debugger"
        description="Real-time step-through debugging with breakpoints, state inspection, and token stream viewer"
      />
      <Suspense fallback={<div className="h-[calc(100vh-12rem)] bg-muted animate-pulse rounded-lg" />}>
        <DebuggerContent />
      </Suspense>
    </div>
  );
}

function DebuggerContent() {
  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="rounded-lg border bg-card p-3 flex items-center gap-2">
        <div className="flex items-center gap-1 border-r pr-3 mr-2">
          <ToolbarButton icon="▶" label="Continue" shortcut="F5" />
          <ToolbarButton icon="⏭" label="Step Over" shortcut="F10" />
          <ToolbarButton icon="⤵" label="Step Into" shortcut="F11" />
          <ToolbarButton icon="⏸" label="Pause" shortcut="F6" />
          <ToolbarButton icon="⏹" label="Stop" shortcut="Shift+F5" />
        </div>
        <div className="flex-1 flex items-center gap-2">
          <input
            type="text"
            placeholder="Enter trace ID or connect to live session..."
            className="flex-1 rounded-md border bg-background px-3 py-1.5 text-sm"
          />
          <button className="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90">
            Connect
          </button>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground border-l pl-3">
          <span className="inline-block w-2 h-2 rounded-full bg-yellow-500" />
          Disconnected
        </div>
      </div>

      {/* Main debug panels - Chrome DevTools inspired layout */}
      <div className="grid gap-4 lg:grid-cols-3" style={{ height: "calc(100vh - 16rem)" }}>
        {/* Left: Trace Timeline */}
        <div className="rounded-lg border bg-card overflow-hidden flex flex-col">
          <div className="border-b p-3 bg-muted/50">
            <h3 className="text-sm font-semibold">Trace Timeline</h3>
          </div>
          <div className="flex-1 overflow-y-auto p-3">
            <div className="text-center py-12 text-muted-foreground">
              <p className="text-sm">No trace loaded</p>
              <p className="text-xs mt-1">Connect to a trace to see the timeline</p>
            </div>
          </div>
        </div>

        {/* Center: Variable Inspector & Token Stream */}
        <div className="rounded-lg border bg-card overflow-hidden flex flex-col">
          <div className="border-b bg-muted/50">
            <div className="flex">
              <TabButton label="Variables" active />
              <TabButton label="Token Stream" active={false} />
              <TabButton label="Output" active={false} />
            </div>
          </div>
          <div className="flex-1 overflow-y-auto p-3">
            <div className="text-center py-12 text-muted-foreground">
              <p className="text-sm">No state to inspect</p>
              <p className="text-xs mt-1">Variables and state will appear here during debugging</p>
            </div>
          </div>
        </div>

        {/* Right: Breakpoints & Cost Meter */}
        <div className="flex flex-col gap-4">
          {/* Breakpoint Manager */}
          <div className="rounded-lg border bg-card overflow-hidden flex-1">
            <div className="border-b p-3 bg-muted/50 flex items-center justify-between">
              <h3 className="text-sm font-semibold">Breakpoints</h3>
              <button className="text-xs rounded border px-2 py-0.5 hover:bg-accent">+ Add</button>
            </div>
            <div className="p-3 space-y-2">
              <BreakpointType icon="🔧" label="On Tool Call" description="Break when a tool is invoked" />
              <BreakpointType icon="💰" label="Cost Threshold" description="Break when cost exceeds limit" />
              <BreakpointType icon="🔍" label="Pattern Match" description="Break on output pattern" />
              <BreakpointType icon="❌" label="On Error" description="Break when error occurs" />
              <BreakpointType icon="📍" label="On Span" description="Break on specific span" />
            </div>
          </div>

          {/* Cost Meter */}
          <div className="rounded-lg border bg-card p-4">
            <h3 className="text-sm font-semibold mb-3">Cost Meter</h3>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Total Cost</span>
                <span className="font-mono font-medium">$0.0000</span>
              </div>
              <div className="w-full bg-muted rounded-full h-2">
                <div className="bg-green-500 h-2 rounded-full" style={{ width: "0%" }} />
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Tokens</span>
                <span className="font-mono font-medium">0</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Elapsed</span>
                <span className="font-mono font-medium">0ms</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Step</span>
                <span className="font-mono font-medium">0/0</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function ToolbarButton({ icon, label, shortcut }: { icon: string; label: string; shortcut: string }) {
  return (
    <button
      className="rounded px-2 py-1 text-lg hover:bg-accent transition-colors"
      title={`${label} (${shortcut})`}
    >
      {icon}
    </button>
  );
}

function TabButton({ label, active }: { label: string; active: boolean }) {
  return (
    <button
      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
        active
          ? "border-primary text-foreground"
          : "border-transparent text-muted-foreground hover:text-foreground"
      }`}
    >
      {label}
    </button>
  );
}

function BreakpointType({ icon, label, description }: { icon: string; label: string; description: string }) {
  return (
    <div className="flex items-center gap-2 p-2 rounded-md hover:bg-muted/50 cursor-pointer">
      <span>{icon}</span>
      <div className="flex-1">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}
