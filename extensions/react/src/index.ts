/**
 * @agenttrace/react - Embeddable React Components for AgentTrace
 *
 * Tree-shakeable component library for embedding trace exploration,
 * cost monitoring, prompt playgrounds, and agent graphs in any React app.
 *
 * @example
 * ```tsx
 * import { TraceViewer, CostChart } from '@agenttrace/react';
 *
 * function App() {
 *   return (
 *     <TraceViewer
 *       apiKey="sk-at-xxx"
 *       traceId="trace-123"
 *       theme="light"
 *     />
 *   );
 * }
 * ```
 */

export { TraceViewer } from "./components/TraceViewer";
export type { TraceViewerProps } from "./components/TraceViewer";

export { CostChart } from "./components/CostChart";
export type { CostChartProps } from "./components/CostChart";

export { PromptPlayground } from "./components/PromptPlayground";
export type { PromptPlaygroundProps } from "./components/PromptPlayground";

export { AgentGraph } from "./components/AgentGraph";
export type { AgentGraphProps } from "./components/AgentGraph";

export { AgentTraceThemeProvider, useAgentTraceTheme } from "./theme/provider";
export type { AgentTraceTheme } from "./theme/provider";
