# @agenttrace/react

Embeddable React component library for AgentTrace. Tree-shakeable, < 50KB per component, no Next.js dependencies.

## Installation

```bash
npm install @agenttrace/react
```

## Components

### TraceViewer
Display trace data with a span tree, timeline, and cost breakdown.

```tsx
import { TraceViewer } from '@agenttrace/react';

<TraceViewer
  apiKey="sk-at-xxx"
  traceId="trace-123"
  theme="light"
  height="600px"
  showCost
  onSpanSelect={(spanId) => console.log(spanId)}
/>
```

### CostChart
Visualize cost breakdowns with a bar chart.

```tsx
import { CostChart } from '@agenttrace/react';

<CostChart
  data={[
    { label: "GPT-4", value: 0.0523 },
    { label: "GPT-3.5", value: 0.0012 },
  ]}
  height={300}
  currency="$"
/>
```

### PromptPlayground
Interactive prompt template editor with variable substitution.

```tsx
import { PromptPlayground } from '@agenttrace/react';

<PromptPlayground
  template="Hello {{name}}, please {{action}}."
  variables={{ name: "Agent", action: "analyze the code" }}
  onRun={(compiled) => console.log(compiled)}
/>
```

### AgentGraph
Visualize agent execution as a directed graph.

```tsx
import { AgentGraph } from '@agenttrace/react';

<AgentGraph
  nodes={[
    { id: "1", label: "Agent", type: "agent" },
    { id: "2", label: "Search", type: "tool" },
  ]}
  edges={[{ source: "1", target: "2" }]}
  onNodeClick={(id) => console.log(id)}
/>
```

## Theming

```tsx
import { AgentTraceThemeProvider } from '@agenttrace/react';

<AgentTraceThemeProvider theme="dark">
  <TraceViewer traceId="..." />
</AgentTraceThemeProvider>
```

CSS variables are also available for Tailwind integration:

```css
:root {
  --at-primary: #3b82f6;
  --at-background: #ffffff;
  --at-foreground: #1f2937;
}
```

## Tree Shaking

Import individual components for minimal bundle size:

```tsx
import { TraceViewer } from '@agenttrace/react/TraceViewer';
import { CostChart } from '@agenttrace/react/CostChart';
```
