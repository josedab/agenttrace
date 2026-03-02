import React, { createContext, useContext } from "react";

export interface AgentTraceTheme {
  mode: "light" | "dark";
  colors: {
    background: string;
    foreground: string;
    border: string;
    muted: string;
    primary: string;
    success: string;
    warning: string;
    error: string;
  };
  fonts: {
    sans: string;
    mono: string;
  };
  radii: {
    sm: string;
    md: string;
    lg: string;
  };
}

const lightTheme: AgentTraceTheme = {
  mode: "light",
  colors: {
    background: "#ffffff",
    foreground: "#1f2937",
    border: "#e5e7eb",
    muted: "#6b7280",
    primary: "#3b82f6",
    success: "#10b981",
    warning: "#f59e0b",
    error: "#ef4444",
  },
  fonts: {
    sans: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    mono: '"SF Mono", "Fira Code", "Fira Mono", Menlo, monospace',
  },
  radii: {
    sm: "4px",
    md: "8px",
    lg: "12px",
  },
};

const darkTheme: AgentTraceTheme = {
  ...lightTheme,
  mode: "dark",
  colors: {
    background: "#1a1a2e",
    foreground: "#e5e7eb",
    border: "#2d2d44",
    muted: "#9ca3af",
    primary: "#60a5fa",
    success: "#34d399",
    warning: "#fbbf24",
    error: "#f87171",
  },
};

const ThemeContext = createContext<AgentTraceTheme>(lightTheme);

export function useAgentTraceTheme(): AgentTraceTheme {
  return useContext(ThemeContext);
}

interface ThemeProviderProps {
  theme?: "light" | "dark" | AgentTraceTheme;
  children: React.ReactNode;
}

export function AgentTraceThemeProvider({ theme = "light", children }: ThemeProviderProps) {
  const resolvedTheme =
    typeof theme === "string"
      ? theme === "dark"
        ? darkTheme
        : lightTheme
      : theme;

  return (
    <ThemeContext.Provider value={resolvedTheme}>
      {children}
    </ThemeContext.Provider>
  );
}

// CSS custom properties for Tailwind integration
export function getThemeCSSVariables(theme: AgentTraceTheme): Record<string, string> {
  return {
    "--at-background": theme.colors.background,
    "--at-foreground": theme.colors.foreground,
    "--at-border": theme.colors.border,
    "--at-muted": theme.colors.muted,
    "--at-primary": theme.colors.primary,
    "--at-success": theme.colors.success,
    "--at-warning": theme.colors.warning,
    "--at-error": theme.colors.error,
    "--at-font-sans": theme.fonts.sans,
    "--at-font-mono": theme.fonts.mono,
    "--at-radius-sm": theme.radii.sm,
    "--at-radius-md": theme.radii.md,
    "--at-radius-lg": theme.radii.lg,
  };
}
