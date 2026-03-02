import React, { useState, useCallback } from "react";

export interface PromptPlaygroundProps {
  /** Prompt template content */
  template: string;
  /** Available variables */
  variables?: Record<string, string>;
  /** Callback when template changes */
  onTemplateChange?: (template: string) => void;
  /** Callback when variables change */
  onVariableChange?: (variables: Record<string, string>) => void;
  /** Callback when user clicks "Run" */
  onRun?: (compiled: string, variables: Record<string, string>) => void;
  /** Theme */
  theme?: "light" | "dark";
  /** Read-only mode */
  readOnly?: boolean;
  /** Custom class name */
  className?: string;
}

export function PromptPlayground({
  template: initialTemplate,
  variables: initialVars = {},
  onTemplateChange,
  onVariableChange,
  onRun,
  theme = "light",
  readOnly = false,
  className = "",
}: PromptPlaygroundProps) {
  const [template, setTemplate] = useState(initialTemplate);
  const [variables, setVariables] = useState<Record<string, string>>(initialVars);

  const isDark = theme === "dark";
  const bg = isDark ? "#1a1a2e" : "#ffffff";
  const border = isDark ? "#2d2d44" : "#e5e7eb";
  const text = isDark ? "#e5e7eb" : "#1f2937";
  const muted = isDark ? "#9ca3af" : "#6b7280";
  const inputBg = isDark ? "#16162a" : "#f9fafb";

  // Extract {{variable}} patterns
  const detectedVars = Array.from(
    new Set(
      (template.match(/\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}/g) || []).map((m) =>
        m.replace(/\{\{|\}\}/g, "")
      )
    )
  );

  const compiled = detectedVars.reduce(
    (result, v) => result.replace(new RegExp(`\\{\\{${v}\\}\\}`, "g"), variables[v] || `{{${v}}}`),
    template
  );

  const handleTemplateChange = useCallback(
    (value: string) => {
      setTemplate(value);
      onTemplateChange?.(value);
    },
    [onTemplateChange]
  );

  const handleVarChange = useCallback(
    (key: string, value: string) => {
      const updated = { ...variables, [key]: value };
      setVariables(updated);
      onVariableChange?.(updated);
    },
    [variables, onVariableChange]
  );

  const handleRun = useCallback(() => {
    onRun?.(compiled, variables);
  }, [compiled, variables, onRun]);

  return (
    <div
      style={{
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        color: text,
        backgroundColor: bg,
        border: `1px solid ${border}`,
        borderRadius: "8px",
        overflow: "hidden",
      }}
      className={className}
    >
      {/* Header */}
      <div
        style={{
          padding: "12px 16px",
          borderBottom: `1px solid ${border}`,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <span style={{ fontWeight: 600, fontSize: "14px" }}>Prompt Playground</span>
        {!readOnly && onRun && (
          <button
            onClick={handleRun}
            style={{
              padding: "6px 16px",
              borderRadius: "6px",
              border: "none",
              backgroundColor: "#3b82f6",
              color: "#fff",
              fontSize: "13px",
              fontWeight: 500,
              cursor: "pointer",
            }}
          >
            ▶ Run
          </button>
        )}
      </div>

      <div style={{ display: "flex" }}>
        {/* Template editor */}
        <div style={{ flex: 2, borderRight: `1px solid ${border}` }}>
          <div style={{ padding: "8px 16px", fontSize: "12px", color: muted, borderBottom: `1px solid ${border}` }}>
            Template
          </div>
          <textarea
            value={template}
            onChange={(e) => handleTemplateChange(e.target.value)}
            readOnly={readOnly}
            style={{
              width: "100%",
              minHeight: "200px",
              padding: "12px 16px",
              border: "none",
              outline: "none",
              resize: "vertical",
              fontFamily: "monospace",
              fontSize: "13px",
              lineHeight: 1.6,
              color: text,
              backgroundColor: bg,
              boxSizing: "border-box",
            }}
          />
        </div>

        {/* Variables panel */}
        <div style={{ flex: 1 }}>
          <div style={{ padding: "8px 16px", fontSize: "12px", color: muted, borderBottom: `1px solid ${border}` }}>
            Variables ({detectedVars.length})
          </div>
          <div style={{ padding: "12px 16px" }}>
            {detectedVars.length === 0 ? (
              <div style={{ color: muted, fontSize: "12px", textAlign: "center", padding: "20px 0" }}>
                No variables detected
              </div>
            ) : (
              detectedVars.map((v) => (
                <div key={v} style={{ marginBottom: "12px" }}>
                  <label style={{ fontSize: "12px", fontWeight: 500, display: "block", marginBottom: "4px" }}>
                    {"{{" + v + "}}"}
                  </label>
                  <input
                    type="text"
                    value={variables[v] || ""}
                    onChange={(e) => handleVarChange(v, e.target.value)}
                    readOnly={readOnly}
                    placeholder={`Enter ${v}...`}
                    style={{
                      width: "100%",
                      padding: "6px 10px",
                      border: `1px solid ${border}`,
                      borderRadius: "4px",
                      fontSize: "13px",
                      color: text,
                      backgroundColor: inputBg,
                      outline: "none",
                      boxSizing: "border-box",
                    }}
                  />
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {/* Preview */}
      <div style={{ borderTop: `1px solid ${border}` }}>
        <div style={{ padding: "8px 16px", fontSize: "12px", color: muted, borderBottom: `1px solid ${border}` }}>
          Compiled Output
        </div>
        <pre
          style={{
            padding: "12px 16px",
            margin: 0,
            fontSize: "13px",
            lineHeight: 1.6,
            fontFamily: "monospace",
            whiteSpace: "pre-wrap",
            color: muted,
            maxHeight: "150px",
            overflow: "auto",
          }}
        >
          {compiled}
        </pre>
      </div>
    </div>
  );
}
