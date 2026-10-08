import { vi } from "vitest";

export const nav = {
  pathname: "/cimory/SNT/dashboard/USR-1/issues",
  params: { plantCode: "SNT", userId: "USR-1" } as Record<string, string>,
};

export const authState = {
  user: { id: "USR-1", name: "Budi Auditor", role_id: "Auditor", plant_id: "SNT" } as Record<string, string> | null,
  logout: vi.fn(),
  setAuth: vi.fn(),
};

vi.mock("next/navigation", () => ({
  usePathname: () => nav.pathname,
  useParams: () => nav.params,
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), prefetch: vi.fn() }),
}));

vi.mock("next/image", () => ({
  default: (props: Record<string, unknown>) => {
    const imgProps = { ...props };
    delete imgProps.priority;
    // eslint-disable-next-line @next/next/no-img-element, jsx-a11y/alt-text
    return <img {...(imgProps as object)} />;
  },
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (selector?: (s: typeof authState) => unknown) => (selector ? selector(authState) : authState),
}));

vi.mock("@/lib/useMounted", () => ({ useMounted: () => true }));
vi.mock("@/lib/usePermissions", () => ({ usePermissions: () => ({ hasPermission: () => true }) }));
