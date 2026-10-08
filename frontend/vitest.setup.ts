import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

// Run tests as a machine outside WIB so date helpers must pin Asia/Jakarta
// themselves instead of leaning on the browser's zone.
process.env.TZ = "UTC";

afterEach(() => cleanup());
