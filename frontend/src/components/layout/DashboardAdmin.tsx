"use client";

import { Factory } from "lucide-react";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { adminWidgetRegistry } from "@/components/dashboard/widgets/admin/registry";
import { AdminDashboardProvider, useAdminDashboard } from "@/components/dashboard/admin/AdminDashboardContext";

function AdminDashboardBody() {
  const { mounted, user, isSuperAdmin, selectedPlant, setSelectedPlant, selectedArea, setSelectedArea, plantsResponse, filteredAreas } =
    useAdminDashboard();

  if (!mounted) {
    return <div className="h-screen w-full bg-background" />;
  }

  return (
    <div className="w-full space-y-8 animate-in fade-in duration-500 pb-24 md:pb-6">
      {/* Header Section */}
      <section className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-foreground">
              Selamat Datang, {user?.name || "Administrator"}
            </h1>
            <span className="bg-primary/10 text-primary text-xs font-semibold px-2.5 py-1 rounded uppercase tracking-wider">
              {isSuperAdmin ? "SUPER ADMIN" : "ADMINISTRATOR"}
            </span>
          </div>
          <p className="text-muted-foreground text-sm">
            Pengawasan Berkelanjutan · Ringkasan aktivitas dan performa seluruh peran
          </p>
        </div>
        <div className="flex flex-wrap gap-2 items-center">
          {isSuperAdmin ? (
            <select
              aria-label="Pilih Plant"
              value={selectedPlant || "all"}
              onChange={(e) => {
                setSelectedPlant(e.target.value);
                setSelectedArea("");
              }}
              className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground font-medium focus:ring-2 focus:ring-primary/40 focus:outline-none min-h-[44px]"
            >
              <option value="all">Semua Plant</option>
              {plantsResponse?.data?.items?.map((plant) => (
                <option key={plant.plant_id} value={plant.plant_id}>
                  {plant.plant_name}
                </option>
              ))}
            </select>
          ) : user?.plant_id ? (
            <div className="px-3 py-1.5 border border-primary/30 bg-primary/5 text-primary rounded-lg text-xs font-semibold uppercase tracking-wider flex items-center gap-1.5">
              <Factory className="h-3.5 w-3.5" />
              Plant: {user.plant_id}
            </div>
          ) : null}

          <select
            aria-label="Pilih Area"
            value={selectedArea}
            onChange={(e) => setSelectedArea(e.target.value)}
            className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground font-medium focus:ring-2 focus:ring-primary/40 focus:outline-none min-h-[44px]"
          >
            <option value="">Semua Area</option>
            {filteredAreas.map((area) => (
              <option key={area.area_id} value={area.area_id}>
                {area.area_name}
              </option>
            ))}
          </select>
        </div>
      </section>

      {/* Widget area — drag to reorder, resize the corner, in Edit mode.
          Falls back to the default arrangement while it loads or if the
          user has never customized it. */}
      <DashboardGrid registry={adminWidgetRegistry} enabled={mounted && !!user} />
    </div>
  );
}

export const DashboardPanelAdmin = () => {
  return (
    <AdminDashboardProvider>
      <AdminDashboardBody />
    </AdminDashboardProvider>
  );
};
