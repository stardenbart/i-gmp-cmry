import { Sidebar } from "@/components/layout/Sidebar";
import { BottomNav } from "@/components/layout/BottomNav";
import { Header } from "@/components/layout/Header";
import { ScopeGuard } from "@/components/layout/ScopeGuard";
import { AppFooter } from "@/components/layout/AppFooter";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <ScopeGuard>
      <div className="flex min-h-dvh w-full bg-background">
        <Header />
        <Sidebar />
        <div className="flex min-w-0 flex-1 flex-col md:pl-[220px]">
          <div aria-hidden="true" className="pwa-header-safe shrink-0" />
          <main className="min-w-0 flex-1 p-3.5 pb-6 sm:p-5">
            <div className="w-full max-w-[1600px] mx-auto">
              {children}
            </div>
          </main>
          <AppFooter />
        </div>
        <BottomNav />
      </div>
    </ScopeGuard>
  );
}
