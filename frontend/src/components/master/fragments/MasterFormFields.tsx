import { useQuery } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import { ChevronDown } from "lucide-react";
import { fetchItems } from "../master.api";
import { useAuthStore } from "@/stores/authStore";
import { isSuperAdminUser } from "@/lib/useAdminGuard";

interface MasterFormFieldsProps {
  activeTab: string;
  editingItem: Record<string, unknown> | null;
}

export function MasterFormFields({ activeTab, editingItem }: MasterFormFieldsProps) {
  const user = useAuthStore((state) => state.user);
  const isSuperAdmin = isSuperAdminUser(user?.role_id);

  // Fetch lookups for forms
  const { data: plantLookup } = useQuery({
    queryKey: ["master", "plants", "lookup"],
    queryFn: () => fetchItems("/master/plants", 1, "", 500),
    enabled: activeTab === "areas" && isSuperAdmin,
  });
  const { data: deptLookup } = useQuery({
    queryKey: ["master", "departments", "lookup"],
    queryFn: () => fetchItems("/master/departments", 1, "", 500),
    enabled: activeTab === "areas",
  });
  const { data: areaLookup } = useQuery({
    queryKey: ["master", "area", "lookup"],
    queryFn: () => fetchItems("/master/area", 1, "", 500),
    enabled: activeTab === "kawasans" || activeTab === "aspeks",
  });
  const { data: kawasanLookup } = useQuery({
    queryKey: ["master", "kawasan", "lookup"],
    queryFn: () => fetchItems("/master/kawasan", 1, "", 500),
    enabled: activeTab === "detail-kawasans" || activeTab === "equipments" || activeTab === "infrastructures",
  });
  const { data: aspekLookup } = useQuery({
    queryKey: ["master", "aspek", "lookup"],
    queryFn: () => fetchItems("/master/aspek", 1, "", 500),
    enabled: activeTab === "details",
  });
  const { data: detailLookup } = useQuery({
    queryKey: ["master", "details", "lookup"],
    queryFn: () => fetchItems("/master/details", 1, "", 500),
    enabled: activeTab === "urains",
  });

  const selectClass = "flex h-12 w-full appearance-none rounded-2xl border border-border bg-card px-4 py-2 pr-10 text-sm text-foreground shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-primary disabled:cursor-not-allowed disabled:opacity-50";

  const fields: Record<string, React.ReactNode> = {
    plants: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kode Plant</label>
          <Input
            name="plant_code"
            defaultValue={editingItem?.plant_code as string}
            placeholder="Contoh: PLT-001"
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Plant (Pabrik)</label>
          <Input
            name="plant_name"
            defaultValue={editingItem?.plant_name as string}
            placeholder="Contoh: Plant Sentul Utama"
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Alamat Plant</label>
          <Input
            name="address"
            defaultValue={editingItem?.address as string}
            placeholder="Contoh: Jl. Industri No. 1, Sentul, Bogor"
          />
        </div>
      </>
    ),
    departments: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Department</label>
          <Input
            name="department_name"
            defaultValue={editingItem?.department_name as string}
            placeholder="Contoh: Quality Assurance"
            required
          />
        </div>
      </>
    ),
    areas: (
      <>
        {isSuperAdmin ? (
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Plant Induk (Pabrik)</label>
            <div className="relative">
              <select name="plant_id" defaultValue={editingItem?.plant_id as string} className={selectClass}>
                <option value="">Pilih Plant (Opsional)...</option>
                {plantLookup?.items?.map((p: any) => (
                  <option key={p.plant_id} value={p.plant_id}>{p.plant_name} ({p.plant_code})</option>
                ))}
              </select>
              <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
            </div>
          </div>
        ) : (
          <input
            type="hidden"
            name="plant_id"
            value={(editingItem?.plant_id as string) || user?.plant_id || ""}
          />
        )}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Department Induk</label>
          <div className="relative">
            <select name="department_id" defaultValue={editingItem?.department_id as string} className={selectClass} required>
              <option value="" disabled>Pilih Department...</option>
              {deptLookup?.items?.map((d: any) => (
                <option key={d.department_id} value={d.department_id}>{d.department_name}</option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Area</label>
          <Input
            name="area_name"
            defaultValue={editingItem?.area_name as string}
            placeholder="Contoh: Pabrik Utama"
            required
          />
        </div>
      </>
    ),
    habits: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kode Habit</label>
          <Input
            name="habit_code"
            defaultValue={editingItem?.habit_code as string}
            placeholder="Contoh: HBT-01"
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Habit</label>
          <Input
            name="habit_name"
            defaultValue={editingItem?.habit_name as string}
            placeholder="Contoh: Penggunaan APD Lengkap"
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kategori Habit</label>
          <Input
            name="habit_category"
            defaultValue={editingItem?.habit_category as string}
            placeholder="Contoh: Safety & Hygiene"
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Deskripsi</label>
          <Input
            name="description"
            defaultValue={editingItem?.description as string}
            placeholder="Keterangan singkat..."
          />
        </div>
      </>
    ),
    equipments: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kawasan Induk</label>
          <div className="relative">
            <select name="kawasan_id" defaultValue={editingItem?.kawasan_id as string} className={selectClass} required>
              <option value="" disabled>Pilih Kawasan...</option>
              {kawasanLookup?.items?.map((k: any) => (
                <option key={k.kawasan_id} value={k.kawasan_id}>{k.kawasan_name || k.name}</option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kode Equipment</label>
          <Input
            name="equipment_code"
            defaultValue={editingItem?.equipment_code as string}
            placeholder="Contoh: EQP-MC-01"
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Equipment</label>
          <Input
            name="equipment_name"
            defaultValue={editingItem?.equipment_name as string}
            placeholder="Contoh: Mesin Pasteurisasi 1"
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Tipe Equipment</label>
          <Input
            name="equipment_type"
            defaultValue={editingItem?.equipment_type as string}
            placeholder="Contoh: Machine"
          />
        </div>
      </>
    ),
    infrastructures: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kawasan Induk</label>
          <div className="relative">
            <select name="kawasan_id" defaultValue={editingItem?.kawasan_id as string} className={selectClass} required>
              <option value="" disabled>Pilih Kawasan...</option>
              {kawasanLookup?.items?.map((k: any) => (
                <option key={k.kawasan_id} value={k.kawasan_id}>{k.kawasan_name || k.name}</option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kode Infrastructure</label>
          <Input
            name="infrastructure_code"
            defaultValue={editingItem?.infrastructure_code as string}
            placeholder="Contoh: INF-DRAIN-01"
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Infrastructure</label>
          <Input
            name="infrastructure_name"
            defaultValue={editingItem?.infrastructure_name as string}
            placeholder="Contoh: Saluran Drainase Utama"
            required
          />
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Tipe Infrastructure</label>
          <Input
            name="infrastructure_type"
            defaultValue={editingItem?.infrastructure_type as string}
            placeholder="Contoh: Facility"
          />
        </div>
      </>
    ),
    kawasans: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Area Induk</label>
          <div className="relative">
            <select name="area_id" defaultValue={editingItem?.area_id as string} className={selectClass} required>
              <option value="" disabled>Pilih Area...</option>
              {areaLookup?.items?.map((a: any) => (
                <option key={a.area_id} value={a.area_id}>{a.area_name}</option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Kawasan</label>
          <Input
            name="kawasan_name"
            defaultValue={editingItem?.kawasan_name as string}
            placeholder="Contoh: Gedung A"
            required
          />
        </div>
      </>
    ),
    "detail-kawasans": (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Kawasan Induk</label>
          <div className="relative">
            <select name="kawasan_id" defaultValue={editingItem?.kawasan_id as string} className={selectClass} required>
              <option value="">Pilih Kawasan...</option>
              {kawasanLookup?.items?.map((k: any) => (
                <option key={k.kawasan_id} value={k.kawasan_id}>
                  {k.kawasan_name || k.name} {k.area?.area_name ? `(Area: ${k.area.area_name})` : k.area_name ? `(Area: ${k.area_name})` : ""}
                </option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Detail Kawasan</label>
          <Input
            name="detail_kawasan_name"
            defaultValue={editingItem?.detail_kawasan_name as string}
            placeholder="Contoh: Lantai 1"
            required
          />
        </div>
      </>
    ),
    aspeks: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Area Induk</label>
          <div className="relative">
            <select name="area_id" defaultValue={editingItem?.area_id as string} className={selectClass} required>
              <option value="" disabled>Pilih Area...</option>
              {areaLookup?.items?.map((a: any) => (
                <option key={a.area_id} value={a.area_id}>{a.area_name}</option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Aspek</label>
          <Input
            name="aspek_name"
            defaultValue={editingItem?.aspek_name as string}
            placeholder="Contoh: Kebersihan"
            required
          />
        </div>
      </>
    ),
    details: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Aspek Induk</label>
          <div className="relative">
            <select name="aspek_id" defaultValue={editingItem?.aspek_id as string} className={selectClass} required>
              <option value="" disabled>Pilih Aspek...</option>
              {aspekLookup?.items?.map((a: any) => (
                <option key={a.aspek_id} value={a.aspek_id}>
                  {a.aspek_name} {a.area?.area_name ? `(Area: ${a.area.area_name})` : ""}
                </option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Nama Detail</label>
          <Input
            name="detail_name"
            defaultValue={editingItem?.detail_name as string}
            placeholder="Contoh: WC Umum"
            required
          />
        </div>
      </>
    ),
    urains: (
      <>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Detail Induk</label>
          <div className="relative">
            <select name="detail_id" defaultValue={editingItem?.detail_id as string} className={selectClass} required>
              <option value="" disabled>Pilih Detail...</option>
              {detailLookup?.items?.map((d: any) => (
                <option key={d.detail_id} value={d.detail_id}>
                  {d.detail_name} {d.aspek?.aspek_name ? `(Aspek: ${d.aspek.aspek_name})` : ""}
                </option>
              ))}
            </select>
            <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Teks Uraian</label>
          <Input
            name="uraian_text"
            defaultValue={editingItem?.uraian_text as string}
            placeholder="Contoh: Lantai bersih dan tidak licin"
            required
          />
        </div>
        <input
          type="hidden"
          name="standard_score"
          value={editingItem ? (editingItem.standard_score as number) : 2}
        />
      </>
    ),
  };

  return <>{fields[activeTab] || null}</>;
}
