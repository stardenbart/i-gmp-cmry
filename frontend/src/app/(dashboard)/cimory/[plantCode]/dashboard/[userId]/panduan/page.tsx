"use client";

import { PanduanFilterProvider } from "@/components/panduan/PanduanFilterContext";
import { RoleFilterBar } from "@/components/panduan/RoleFilterBar";
import { ChapterTabs, type ChapterItem } from "@/components/panduan/ChapterTabs";
import { WelcomeSection } from "@/components/panduan/sections/WelcomeSection";
import { MulaiSection } from "@/components/panduan/sections/MulaiSection";
import { DashboardSection } from "@/components/panduan/sections/DashboardSection";
import { InspeksiSection } from "@/components/panduan/sections/InspeksiSection";
import { TemuanSection } from "@/components/panduan/sections/TemuanSection";
import { WowrSection } from "@/components/panduan/sections/WowrSection";
import { GmpSection } from "@/components/panduan/sections/GmpSection";
import { MonitoringSection } from "@/components/panduan/sections/MonitoringSection";
import { AdminSection } from "@/components/panduan/sections/AdminSection";
import { ReferensiSection } from "@/components/panduan/sections/ReferensiSection";

const CHAPTERS: ChapterItem[] = [
  { id: "selamat-datang", label: "Selamat Datang" },
  { id: "memulai", label: "Memulai" },
  { id: "dashboard", label: "Dashboard" },
  { id: "inspeksi", label: "Inspeksi" },
  { id: "temuan", label: "Temuan Inspeksi" },
  { id: "wowr", label: "Perintah Kerja" },
  { id: "gmp", label: "Data Inspeksi (GMP)" },
  { id: "monitoring", label: "Monitoring" },
  { id: "admin", label: "Administrasi" },
  { id: "referensi", label: "Referensi" },
];

export default function PanduanPage() {
  return (
    <PanduanFilterProvider>
      <div className="space-y-10 pb-16">
        <div className="sticky top-16 z-20 -mx-3.5 space-y-3 border-b border-border bg-background/95 px-3.5 py-3 shadow-sm backdrop-blur-xl sm:-mx-6 sm:px-6">
          <RoleFilterBar />
          <ChapterTabs chapters={CHAPTERS} />
        </div>

        <div className="mx-auto max-w-4xl space-y-14">
          <WelcomeSection />
          <MulaiSection />
          <DashboardSection />
          <InspeksiSection />
          <TemuanSection />
          <WowrSection />
          <GmpSection />
          <MonitoringSection />
          <AdminSection />
          <ReferensiSection />
        </div>
      </div>
    </PanduanFilterProvider>
  );
}
