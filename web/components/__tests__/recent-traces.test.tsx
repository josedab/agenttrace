import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { RecentTraces } from "../dashboard/recent-traces";

// Mock next/link
vi.mock("next/link", () => ({
  default: ({ children, href, ...props }: { children: React.ReactNode; href: string }) => {
    return <a href={href} {...props}>{children}</a>;
  },
}));

// Mock date-fns to avoid time-dependent test flakiness
vi.mock("date-fns", () => ({
  formatDistanceToNow: () => "2 hours ago",
}));

import { vi } from "vitest";
import * as React from "react";

const mockTraces = [
  {
    id: "trace-1",
    name: "API Request",
    input: "What is the capital of France?",
    output: "Paris is the capital of France.",
    startTime: "2026-02-20T10:00:00Z",
    latency: 1500,
    totalCost: 0.0042,
    level: "DEFAULT" as const,
  },
  {
    id: "trace-2",
    name: null,
    input: null,
    output: null,
    startTime: "2026-02-20T09:00:00Z",
    latency: null,
    totalCost: null,
    level: "ERROR" as const,
  },
];

describe("RecentTraces", () => {
  it("renders empty state when no traces", () => {
    render(<RecentTraces traces={[]} />);

    expect(screen.getByText("No traces yet")).toBeInTheDocument();
    expect(
      screen.getByText("Start instrumenting your agents to see traces here")
    ).toBeInTheDocument();
  });

  it("renders the component title", () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("Recent Traces")).toBeInTheDocument();
    expect(screen.getByText("Latest agent executions")).toBeInTheDocument();
  });

  it("renders trace name when available", () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("API Request")).toBeInTheDocument();
  });

  it("renders truncated trace ID when name is null", () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("trace-2")).toBeInTheDocument();
  });

  it("renders input text truncated", () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("What is the capital of France?")).toBeInTheDocument();
  });

  it('renders "No input" when input is null', () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("No input")).toBeInTheDocument();
  });

  it("renders level badges", () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("DEFAULT")).toBeInTheDocument();
    expect(screen.getByText("ERROR")).toBeInTheDocument();
  });

  it("renders latency when available", () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("1.50s")).toBeInTheDocument();
  });

  it("renders cost when available", () => {
    render(<RecentTraces traces={mockTraces} />);

    expect(screen.getByText("$0.0042")).toBeInTheDocument();
  });

  it('renders "View all" link', () => {
    render(<RecentTraces traces={mockTraces} />);

    const viewAllLink = screen.getByText("View all");
    expect(viewAllLink.closest("a")).toHaveAttribute("href", "/traces");
  });

  it("links each trace to its detail page", () => {
    render(<RecentTraces traces={mockTraces} />);

    const links = screen.getAllByRole("link");
    const traceLinks = links.filter(
      (link) =>
        link.getAttribute("href")?.startsWith("/traces/")
    );
    expect(traceLinks).toHaveLength(2);
    expect(traceLinks[0]).toHaveAttribute("href", "/traces/trace-1");
    expect(traceLinks[1]).toHaveAttribute("href", "/traces/trace-2");
  });
});
