import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import "@/test/shell-mocks";
import { NotificationBell } from "./NotificationBell";

vi.mock("@/lib/api/notification.api", () => ({
  notificationApi: {
    getAll: vi.fn().mockResolvedValue({ items: [], total: 0 }),
    getNotifications: vi.fn().mockResolvedValue({ items: [], total: 0 }),
    markAsRead: vi.fn(),
    markAllAsRead: vi.fn(),
  },
}));
vi.mock("@/hooks/usePushNotifications", () => ({
  usePushNotifications: () => ({ state: "default", isSubscribed: false, isBusy: false, subscribe: vi.fn(), unsubscribe: vi.fn() }),
}));

describe("NotificationBell", () => {
  // The bell lives in the brand header, which sets text-white. The panel
  // must reset to the card's text colour or unread titles turn white on the
  // white card in light mode.
  it("renders the dropdown in the card text colour, not the header's white", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <div className="text-white">
          <NotificationBell triggerClassName="text-white" />
        </div>
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: /Pusat Notifikasi/ }));

    const panel = screen.getByRole("dialog", { name: "Daftar Notifikasi" });
    expect(panel).toHaveClass("bg-card", "text-card-foreground");
  });
});
