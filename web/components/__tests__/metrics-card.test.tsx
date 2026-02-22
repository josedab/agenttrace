import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MetricsCard } from "../dashboard/metrics-card";
import { Activity } from "lucide-react";

describe("MetricsCard", () => {
  it("renders title and value", () => {
    render(
      <MetricsCard title="Total Traces" value="1.5K" icon={Activity} />
    );

    expect(screen.getByText("Total Traces")).toBeInTheDocument();
    expect(screen.getByText("1.5K")).toBeInTheDocument();
  });

  it("renders trend indicator when change is provided", () => {
    render(
      <MetricsCard
        title="Active Sessions"
        value="42"
        change={12.5}
        icon={Activity}
        trend="up"
      />
    );

    expect(screen.getByText("12.5%")).toBeInTheDocument();
    expect(screen.getByText("from last period")).toBeInTheDocument();
  });

  it("shows positive trend styling for 'up' trend", () => {
    render(
      <MetricsCard
        title="Traces"
        value="100"
        change={5.0}
        icon={Activity}
        trend="up"
      />
    );

    const changeElement = screen.getByText("5.0%");
    expect(changeElement).toHaveClass("text-green-500");
  });

  it("shows negative trend styling for 'down' trend", () => {
    render(
      <MetricsCard
        title="Cost"
        value="$42.00"
        change={-3.2}
        icon={Activity}
        trend="down"
      />
    );

    const changeElement = screen.getByText("3.2%");
    expect(changeElement).toHaveClass("text-red-500");
  });

  it("does not render change section when change is undefined", () => {
    render(
      <MetricsCard title="Latency" value="250ms" icon={Activity} />
    );

    expect(screen.queryByText("from last period")).not.toBeInTheDocument();
  });

  it("applies custom className", () => {
    const { container } = render(
      <MetricsCard
        title="Test"
        value="0"
        icon={Activity}
        className="custom-class"
      />
    );

    const card = container.firstChild as HTMLElement;
    expect(card.className).toContain("custom-class");
  });

  it("renders with zero change value", () => {
    render(
      <MetricsCard
        title="Stable"
        value="50"
        change={0}
        icon={Activity}
        trend="up"
      />
    );

    expect(screen.getByText("0.0%")).toBeInTheDocument();
  });
});
