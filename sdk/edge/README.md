# AgentTrace Edge SDK

Lightweight (<50KB) SDK for monitoring AI agents running on mobile devices, IoT, and edge environments.

## Features

- **Offline trace buffering**: Automatically buffers events when offline and syncs when connectivity is restored
- **Bandwidth-optimized batching**: Groups events into compressed batches to minimize network usage
- **Privacy-preserving local aggregation**: Aggregates metrics locally before uploading to reduce data exposure
- **Multi-platform**: Swift (iOS), Kotlin (Android), WASM (Web/Edge), C (IoT)

## Quick Start

### JavaScript/WASM

```javascript
import { AgentTraceEdge } from '@agenttrace/edge-sdk';

const tracer = new AgentTraceEdge({
  apiUrl: 'https://your-instance.agenttrace.io',
  apiKey: 'at_...',
  deviceId: 'device-001',
  platform: 'wasm',
  batchSize: 50,
  flushIntervalMs: 30000,
  offlineBufferSize: 1000,
});

// Record an event
tracer.recordEvent({
  type: 'trace',
  name: 'user-query',
  input: 'What is the weather?',
  output: 'It is sunny today.',
  model: 'gpt-4o-mini',
  latencyMs: 450,
  tokensIn: 12,
  tokensOut: 8,
});

// Force flush
await tracer.flush();
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/public/edge/devices` | Register a device |
| POST | `/api/public/edge/ingest` | Ingest a batch of events |
| GET | `/api/public/edge/devices` | List registered devices |
| GET | `/api/public/edge/devices/:id/status` | Get device status |
| POST | `/api/public/edge/sync` | Sync offline data |
| GET | `/api/public/edge/stats` | Get edge statistics |

## Architecture

```
Edge Device                    AgentTrace Server
┌─────────────┐               ┌──────────────────┐
│ Agent Code  │               │                  │
│      │      │               │  Edge Ingest API │
│  Edge SDK   │──batch POST──▶│       │          │
│      │      │               │  Trace Storage   │
│ Local Buffer│               │       │          │
│ (offline)   │               │  Analytics       │
└─────────────┘               └──────────────────┘
```
