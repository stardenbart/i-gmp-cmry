import { Sidebar } from "@/components/layout/Sidebar";
import { BottomNav } from "@/components/layout/BottomNav";
import { Header } from "@/components/layout/Header";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen w-full bg-background">
      <Sidebar />
      <div className="flex flex-1 flex-col md:pl-64">
        <Header />
        <main className="flex-1 p-4 sm:p-6 pb-20 md:pb-6">
          <div className="mx-auto max-w-6xl w-full">
            {children}
          </div>
        </main>
      </div>
      <BottomNav />
    </div>
  );
}
