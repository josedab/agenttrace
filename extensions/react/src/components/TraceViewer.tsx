import React, { useState, useEffect, useCallback } from "react";

export interface TraceViewerProps {
  /** AgentTrace API key for data fetching */
  apiKey?: string;
  /** Base URL of the AgentTrace API */
  apiUrl?: string;
  /** Trace ID to display */
  traceId: string;
  /** Visual theme */
  theme?: "light" | "dark" | "system";
  /** Component height */
  height?: string | number;
  /** Show timeline panel */
  showTimeline?: boolean;
  /** Show cost breakdown */
  showCost?: boolean;
  /** Custom class name */
  className?: string;
  /** Callback when a span is selected */
  onSpanSelect?: (spanId: string) => void;
  /** Initial data (skip API fetch) */
  data?: TraceData;
}

export interface TraceData {
  id: string;
  name: string;
  startTime: string;
  endTime?: string;
  duration?: number;
  totalCost?: number;
  totalTokens?: number;
  spans: SpanData[];
}

export interface SpanData {
  id: string;
  name: string;
  type: string;
  parentId?: string;
  startTime: string;
  endTime?: string;
  duration?: number;
  cost?: number;
  tokens?: number;
  input?: unknown;
  output?: unknown;
  level?: string;
  metadata?: Record<string, string>;
}

export function TraceViewer({
  apiKey,
  apiUrl = "https://api.agenttrace.io",
  traceId,
  theme = "light",
  height = "600px",
  showTimeline = true,
  showCost = true,
  className = "",
  onSpanSelect,
  data: initialData,
}: TraceViewerProps) {
  const [data, setData] = useState<TraceData | null>(initialData ?? null);
  const [selectedSpan, setSelectedSpan] = useState<string | null>(null);
  const [loading, setLoading] = useState(!initialData);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (initialData || !apiKey) return;

    const fetchTrace = async () => {
      try {
        setLoading(true);
        const resp = await fetch(`${apiUrl}/api/public/traces/${traceId}`, {
          headers: { Authorization: `Bearer ${apiKey}` },
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const json = await resp.json();
        setData(json);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load trace");
      } finally {
        setLoading(false);
      }
    };

    fetchTrace();
  }, [apiKey, apiUrl, traceId, initialData]);

  const handleSpanClick = useCallback(
    (spanId: string) => {
      setSelectedSpan(spanId);
      onSpanSelect?.(spanId);
    },
    [onSpanSelect]
  );

  const isDark = theme === "dark";
  const styles = getStyles(isDark);

  if (loading) {
    return (
      <div style={{ ...styles.container, height }} className={className}>
        <div style={styles.loading}>Loading trace...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ ...styles.container, height }} className={className}>
        <div style={styles.error}>Error: {error}</div>
      </div>
    );
  }

  if (!data) {
    return (
      <div style={{ ...styles.container, height }} className={className}>
        <div style={styles.empty}>No trace data</div>
      </div>
    );
  }

  const rootSpans = data.spans.filter((s) => !s.parentId);

  return (
    <div style={{ ...styles.container, height }} className={className}>
      {/* Header */}
      <div style={styles.header}>
        <div>
          <div style={styles.traceName}>{data.name}</div>
          <div style={styles.traceId}>{data.id}</div>
        </div>
        <div style={styles.headerStats}>
          {data.duration != null && (
            <span style={styles.stat}>{data.duration}ms</span>
          )}
          {showCost && data.totalCost != null && (
            <span style={styles.stat}>${data.totalCost.toFixed(4)}</span>
          )}
          {data.totalTokens != null && (
            <span style={styles.stat}>{data.totalTokens} tokens</span>
          )}
        </div>
      </div>

      {/* Span tree */}
      <div style={styles.spanTree}>
        {rootSpans.map((span) => (
          <SpanNode
            key={span.id}
            span={span}
            allSpans={data.spans}
            depth={0}
            selectedId={selectedSpan}
            onClick={handleSpanClick}
            isDark={isDark}
          />
        ))}
      </div>

      {/* Selected span detail */}
      {selectedSpan && (
        <SpanDetail
          span={data.spans.find((s) => s.id === selectedSpan)}
          isDark={isDark}
        />
      )}
    </div>
  );
}

function SpanNode({
  span,
  allSpans,
  depth,
  selectedId,
  onClick,
  isDark,
}: {
  span: SpanData;
  allSpans: SpanData[];
  depth: number;
  selectedId: string | null;
  onClick: (id: string) => void;
  isDark: boolean;
}) {
  const children = allSpans.filter((s) => s.parentId === span.id);
  const isSelected = selectedId === span.id;
  const styles = getStyles(isDark);

  return (
    <div>
      <div
        style={{
          ...styles.spanRow,
          paddingLeft: `${depth * 20 + 8}px`,
          backgroundColor: isSelected
            ? isDark
              ? "#1e3a5f"
              : "#e0f2fe"
            : "transparent",
          cursor: "pointer",
        }}
        onClick={() => onClick(span.id)}
      >
        <span style={styles.spanType}>{span.type}</span>
        <span style={styles.spanName}>{span.name}</span>
        <span style={styles.spanDuration}>
          {span.duration != null ? `${span.duration}ms` : ""}
        </span>
      </div>
      {children.map((child) => (
        <SpanNode
          key={child.id}
          span={child}
          allSpans={allSpans}
          depth={depth + 1}
          selectedId={selectedId}
          onClick={onClick}
          isDark={isDark}
        />
      ))}
    </div>
  );
}

function SpanDetail({
  span,
  isDark,
}: {
  span?: SpanData;
  isDark: boolean;
}) {
  if (!span) return null;
  const styles = getStyles(isDark);

  return (
    <div style={styles.detailPanel}>
      <div style={styles.detailHeader}>{span.name}</div>
      <div style={styles.detailRow}>
        <span>Type:</span>
        <span>{span.type}</span>
      </div>
      {span.duration != null && (
        <div style={styles.detailRow}>
          <span>Duration:</span>
          <span>{span.duration}ms</span>
        </div>
      )}
      {span.cost != null && (
        <div style={styles.detailRow}>
          <span>Cost:</span>
          <span>${span.cost.toFixed(4)}</span>
        </div>
      )}
      {span.tokens != null && (
        <div style={styles.detailRow}>
          <span>Tokens:</span>
          <span>{span.tokens}</span>
        </div>
      )}
    </div>
  );
}

function getStyles(isDark: boolean) {
  const bg = isDark ? "#1a1a2e" : "#ffffff";
  const border = isDark ? "#2d2d44" : "#e5e7eb";
  const text = isDark ? "#e5e7eb" : "#1f2937";
  const muted = isDark ? "#9ca3af" : "#6b7280";

  return {
    container: {
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
      fontSize: "14px",
      color: text,
      backgroundColor: bg,
      border: `1px solid ${border}`,
      borderRadius: "8px",
      overflow: "hidden" as const,
      display: "flex" as const,
      flexDirection: "column" as const,
    },
    header: {
      padding: "12px 16px",
      borderBottom: `1px solid ${border}`,
      display: "flex" as const,
      justifyContent: "space-between" as const,
      alignItems: "center" as const,
    },
    traceName: { fontWeight: 600 as const, fontSize: "16px" },
    traceId: { color: muted, fontSize: "12px", marginTop: "2px" },
    headerStats: { display: "flex" as const, gap: "12px" },
    stat: {
      fontSize: "13px",
      color: muted,
      fontFamily: "monospace",
    },
    spanTree: { flex: 1, overflowY: "auto" as const },
    spanRow: {
      padding: "6px 8px",
      display: "flex" as const,
      gap: "8px",
      alignItems: "center" as const,
      borderBottom: `1px solid ${border}`,
      fontSize: "13px",
    },
    spanType: {
      fontSize: "11px",
      padding: "1px 6px",
      borderRadius: "4px",
      backgroundColor: isDark ? "#2d2d44" : "#f3f4f6",
      color: muted,
    },
    spanName: { flex: 1 },
    spanDuration: { color: muted, fontFamily: "monospace", fontSize: "12px" },
    detailPanel: {
      padding: "12px 16px",
      borderTop: `1px solid ${border}`,
      backgroundColor: isDark ? "#16162a" : "#f9fafb",
    },
    detailHeader: { fontWeight: 600 as const, marginBottom: "8px" },
    detailRow: {
      display: "flex" as const,
      justifyContent: "space-between" as const,
      padding: "4px 0",
      fontSize: "13px",
    },
    loading: { padding: "40px", textAlign: "center" as const, color: muted },
    error: { padding: "40px", textAlign: "center" as const, color: "#ef4444" },
    empty: { padding: "40px", textAlign: "center" as const, color: muted },
  };
}
