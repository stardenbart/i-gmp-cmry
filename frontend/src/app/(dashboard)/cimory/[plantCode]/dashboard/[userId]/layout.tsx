import { Sidebar } from "@/components/layout/Sidebar";
import { BottomNav } from "@/components/layout/BottomNav";
import { Header } from "@/components/layout/Header";
import { ScopeGuard } from "@/components/layout/ScopeGuard";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <ScopeGuard>
      <div className="flex min-h-dvh w-full bg-background">
        <Sidebar />
        <div className="flex flex-1 flex-col md:pl-64 min-w-0">
          <Header />
          <main className="pwa-content-safe flex-1 p-3.5 sm:p-6 min-w-0">
            <div className="w-full max-w-[1600px] mx-auto">
              {children}
            </div>
          </main>
        </div>
        <BottomNav />
      </div>
    </ScopeGuard>
  );
}
