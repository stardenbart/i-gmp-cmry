"use client";

import { useState } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useMounted } from "@/lib/useMounted";
import { cn } from "@/lib/utils";

const MOCK_INSPECTORS = [
  {
    initials: "BS",
    name: "Budi Santoso",
    role: "Lead Inspector",
    score: 88,
    selesai: 12,
    berjalan: 4,
    belum: 2,
  },
  {
    initials: "SA",
    name: "Siti Aminah",
    role: "Senior Auditor",
    score: 92,
    selesai: 18,
    berjalan: 2,
    belum: 0,
  }
];

export default function MonitoringPage() {
  const mounted = useMounted();
  const user = useAuthStore((state) => state.user);
  const [activeTab, setActiveTab] = useState<"audit" | "pic" | "auditee">("audit");

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  return (
    <div className="w-full space-y-6 animate-in fade-in duration-500 pb-24 md:pb-6">
      
      {/* Header Section */}
      <section className="flex flex-col gap-1 mb-2">
        <h2 className="text-3xl font-bold tracking-tight text-primary">Monitoring</h2>
        <p className="text-muted-foreground text-sm">
          Overview of current audit activities and performance metrics.
        </p>
      </section>

      {/* Tabs Navigation */}
      <nav className="flex border-b border-border w-full overflow-x-auto hide-scrollbar sticky top-[64px] bg-background z-30 pt-2 pb-0">
        <button 
          className={cn(
            "flex-1 py-2 px-4 text-center font-semibold border-b-2 transition-colors",
            activeTab === "audit" ? "text-primary border-primary" : "text-muted-foreground border-transparent hover:text-primary"
          )}
          onClick={() => setActiveTab("audit")}
        >
          Audit
        </button>
        <button 
          className={cn(
            "flex-1 py-2 px-4 text-center font-semibold border-b-2 transition-colors",
            activeTab === "pic" ? "text-primary border-primary" : "text-muted-foreground border-transparent hover:text-primary"
          )}
          onClick={() => setActiveTab("pic")}
        >
          PIC
        </button>
        <button 
          className={cn(
            "flex-1 py-2 px-4 text-center font-semibold border-b-2 transition-colors",
            activeTab === "auditee" ? "text-primary border-primary" : "text-muted-foreground border-transparent hover:text-primary"
          )}
          onClick={() => setActiveTab("auditee")}
        >
          Auditee
        </button>
      </nav>

      {/* TAB CONTENT: AUDIT */}
      {activeTab === "audit" && (
        <div className="flex flex-col gap-4 mt-2 animate-in slide-in-from-bottom-2">
          {/* Bento Stats Row */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-card border border-border rounded-lg p-4 shadow-sm flex flex-col gap-1">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Total Active</span>
              <span className="text-3xl font-bold text-primary">24</span>
            </div>
            <div className="bg-card border border-border rounded-lg p-4 shadow-sm flex flex-col gap-1">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Avg Compliance</span>
              <span className="text-3xl font-bold text-green-500">86%</span>
            </div>
          </div>
          
          <h3 className="text-lg font-bold text-primary mt-2 border-b border-border pb-1">Inspectors</h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {MOCK_INSPECTORS.map((inspector, idx) => (
              <div key={idx} className="bg-card border border-border rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow cursor-pointer flex flex-col group">
                <div className="p-4 flex justify-between items-start border-b border-border bg-muted/30 group-hover:bg-muted/50 transition-colors">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold">
                      {inspector.initials}
                    </div>
                    <div className="flex flex-col">
                      <span className="font-semibold text-foreground">{inspector.name}</span>
                      <span className="text-xs text-muted-foreground">{inspector.role}</span>
                    </div>
                  </div>
                  <div className="flex flex-col items-end">
                    <span className="font-bold text-green-500 text-lg">{inspector.score}%</span>
                    <span className="text-[10px] font-semibold text-muted-foreground uppercase">Score</span>
                  </div>
                </div>
                <div className="p-4 flex flex-col gap-2 grow bg-card">
                  <div className="flex justify-between items-center w-full">
                    <div className="flex flex-col items-center flex-1">
                      <span className="text-lg font-bold text-green-500">{inspector.selesai}</span>
                      <span className="text-[10px] font-semibold text-muted-foreground uppercase">Selesai</span>
                    </div>
                    <div className="w-px h-8 bg-border"></div>
                    <div className="flex flex-col items-center flex-1">
                      <span className="text-lg font-bold text-primary">{inspector.berjalan}</span>
                      <span className="text-[10px] font-semibold text-muted-foreground uppercase">Berjalan</span>
                    </div>
                    <div className="w-px h-8 bg-border"></div>
                    <div className="flex flex-col items-center flex-1">
                      <span className="text-lg font-bold text-muted-foreground">{inspector.belum}</span>
                      <span className="text-[10px] font-semibold text-muted-foreground uppercase">Belum</span>
                    </div>
                  </div>
                </div>
                <div className="p-2 px-4 bg-card border-t border-border mt-auto">
                  <button className="w-full py-2 text-xs font-semibold border border-primary text-primary rounded-md hover:bg-primary/10 transition-colors">
                    Detail Aktivitas
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB CONTENT: PIC */}
      {activeTab === "pic" && (
        <div className="flex flex-col gap-4 mt-2 animate-in slide-in-from-bottom-2">
          <h3 className="text-lg font-bold text-primary mt-2 border-b border-border pb-1">Person in Charge (PIC)</h3>
          <p className="text-sm text-muted-foreground">Coming soon...</p>
        </div>
      )}

      {/* TAB CONTENT: AUDITEE */}
      {activeTab === "auditee" && (
        <div className="flex flex-col gap-4 mt-2 animate-in slide-in-from-bottom-2">
          <h3 className="text-lg font-bold text-primary mt-2 border-b border-border pb-1">Auditee Overview</h3>
          <p className="text-sm text-muted-foreground">Coming soon...</p>
        </div>
      )}
    </div>
  );
}
