import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StatsCards, type DashboardStats } from "./StatCard";

const stats: DashboardStats = {
  total_inspections_running: 1,
  compliance_rate: 80,
  compliance_rate_trend: 0,
  total_issues: 6,
  total_open_issues: 4,
  issue_overdue: 2,
  issue_overdue_trend: 0,
  issues_resolved: 5,
};

describe("StatsCards", () => {
  it("shows a Temuan Closed card instead of Temuan Jatuh Tempo", () => {
    render(<StatsCards stats={stats} isLoading={false} />);

    expect(screen.queryByText("Temuan Jatuh Tempo")).toBeNull();
    const card = screen.getByText("Temuan Closed").parentElement!;
    expect(card).toHaveTextContent("5");
  });
});
