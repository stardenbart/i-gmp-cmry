import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { AdminDashboardStats } from "@/components/dashboard/admin/AdminDashboardContext";
import { AktivitasInspeksiWidget } from "./AktivitasInspeksiWidget";
import { AktivitasPICWidget } from "./AktivitasPICWidget";
import { StatusAuditeeWidget } from "./StatusAuditeeWidget";
import { StatusWOWRWidget } from "./StatusWOWRWidget";

const state: { stats: Partial<AdminDashboardStats> } = { stats: {} };

vi.mock("@/components/dashboard/admin/AdminDashboardContext", () => ({
  useAdminDashboard: () => ({ stats: state.stats, user: { id: "USR-1", plant_id: "PLT-SENTUL" } }),
}));

beforeEach(() => {
  state.stats = {};
});

describe("AktivitasInspeksiWidget", () => {
  it("stacks completed and running inspections in one bar of the real total", () => {
    state.stats = { inspections_completed: 3, inspections_running: 1 };
    render(<AktivitasInspeksiWidget />);

    expect(screen.getByText("Total 4")).toBeInTheDocument();
    const bar = screen.getByRole("img", { name: /Selesai 3 dari 4, Berlangsung 1 dari 4/ });
    expect(within(bar).getByTestId("segment-completed")).toHaveStyle({ width: "75%" });
    expect(within(bar).getByTestId("segment-running")).toHaveStyle({ width: "25%" });
    expect(screen.queryByText(/3 \/ 4/)).toBeNull();
  });
});

describe("AktivitasPICWidget", () => {
  it("splits all findings into resolved, open and overdue in one bar", () => {
    // 5 findings: 2 closed, 3 open of which 1 is past due.
    state.stats = { issues_resolved: 2, total_open_issues: 3, issue_overdue: 1 };
    render(<AktivitasPICWidget />);

    expect(screen.getByText("Total 5 Temuan")).toBeInTheDocument();
    const bar = screen.getByRole("img", { name: /Diselesaikan 2 dari 5, Terbuka 2 dari 5, Jatuh Tempo 1 dari 5/ });
    expect(within(bar).getByTestId("segment-resolved")).toHaveStyle({ width: "40%" });
    expect(within(bar).getByTestId("segment-open")).toHaveStyle({ width: "40%" });
    expect(within(bar).getByTestId("segment-overdue")).toHaveStyle({ width: "20%" });
  });
});

describe("StatusAuditeeWidget", () => {
  it("shows only the area name, without the area-type subtitle", () => {
    state.stats = {
      auditee_status: [{ name: "Area Produksi", type: "Area Operasional", compliance: 80, open_issues: 1 }],
    };
    render(<StatusAuditeeWidget />);

    expect(screen.getByText("Area Produksi")).toBeInTheDocument();
    expect(screen.queryByText("Area Operasional")).toBeNull();
  });
});

describe("StatusWOWRWidget", () => {
  it("stacks approved, waiting, rejected and no-proof WO/WR and keeps their counts", () => {
    state.stats = { wowr_total: 10, wowr_verified: 4, wowr_pending: 3, wowr_rejected: 1, wowr_awaiting: 2, wowr_verified_rate: 40 };
    render(<StatusWOWRWidget />);

    const bar = screen.getByRole("img", { name: /Disetujui 4 dari 10, Menunggu 3 dari 10, Ditolak 1 dari 10, Belum Bukti 2 dari 10/ });
    expect(within(bar).getByTestId("segment-verified")).toHaveStyle({ width: "40%" });

    const legend = screen.getByRole("list");
    for (const [label, count] of [["Menunggu", "3"], ["Ditolak", "1"], ["Belum Bukti", "2"]]) {
      expect(within(legend).getByText(label).closest("li")).toHaveTextContent(count);
    }
    expect(screen.getByText("Tingkat Verifikasi: 40%")).toBeInTheDocument();
  });
});
