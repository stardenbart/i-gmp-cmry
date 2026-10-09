import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AspekStatusPanel } from "./AspekStatusPanel";

vi.mock("@/lib/api/inspection.api", () => ({
  inspectionApi: {
    getKawasanStatus: vi.fn().mockResolvedValue({
      data: [
        { aspek_id: "A2", status: "LOCKED", locked_by: "USR-OTHER-9" },
        { aspek_id: "A3", status: "LOCKED", locked_by: "USR-ME" },
      ],
    }),
  },
}));
vi.mock("@/hooks/useRealtimeSync", () => ({ useRealtimeSync: () => ({ isWebSocketActive: true }) }));

describe("AspekStatusPanel", () => {
  it("labels aspek access as Unlocked / Locked", async () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <AspekStatusPanel
          kawasanId="KWS-1"
          currentUserId="USR-ME"
          aspeks={[
            { aspek_id: "A1", aspek_name: "Lingkungan" },
            { aspek_id: "A2", aspek_name: "Konstruksi" },
            { aspek_id: "A3", aspek_name: "Area Produksi" },
          ]}
        />
      </QueryClientProvider>,
    );

    expect(await screen.findByText("Locked (USR-OT)")).toBeInTheDocument();
    expect(screen.getByText("Unlocked")).toBeInTheDocument();
    expect(screen.getByText("Aktif Mengedit")).toBeInTheDocument();
    expect(screen.queryByText("FREE")).toBeNull();
  });
});
