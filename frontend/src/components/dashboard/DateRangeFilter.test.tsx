import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { DateRangeFilter } from "./DateRangeFilter";

describe("DateRangeFilter", () => {
  it("reports the picked start and end dates", () => {
    const onChange = vi.fn();
    const { rerender } = render(<DateRangeFilter value={{ start: "", end: "" }} onChange={onChange} />);

    fireEvent.change(screen.getByLabelText("Dari tanggal"), { target: { value: "2026-10-01" } });
    expect(onChange).toHaveBeenLastCalledWith({ start: "2026-10-01", end: "" });

    rerender(<DateRangeFilter value={{ start: "2026-10-01", end: "" }} onChange={onChange} />);
    fireEvent.change(screen.getByLabelText("Sampai tanggal"), { target: { value: "2026-10-08" } });
    expect(onChange).toHaveBeenLastCalledWith({ start: "2026-10-01", end: "2026-10-08" });
  });

  it("keeps the end date from going before the start date", () => {
    render(<DateRangeFilter value={{ start: "2026-10-05", end: "2026-10-08" }} onChange={vi.fn()} />);
    expect(screen.getByLabelText("Sampai tanggal")).toHaveAttribute("min", "2026-10-05");
    expect(screen.getByLabelText("Dari tanggal")).toHaveAttribute("max", "2026-10-08");
  });

  it("offers a reset back to all time only when a date is set", () => {
    const onChange = vi.fn();
    const { rerender } = render(<DateRangeFilter value={{ start: "", end: "" }} onChange={onChange} />);
    expect(screen.queryByRole("button", { name: /Semua waktu/ })).toBeNull();

    rerender(<DateRangeFilter value={{ start: "2026-10-01", end: "" }} onChange={onChange} />);
    fireEvent.click(screen.getByRole("button", { name: /Semua waktu/ }));
    expect(onChange).toHaveBeenLastCalledWith({ start: "", end: "" });
  });
});
