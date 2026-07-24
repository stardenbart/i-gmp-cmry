import { useQuery } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import { ChevronDown } from "lucide-react";
import { fetchItems } from "../master.api";

interface MasterFormFieldsProps {
  activeTab: string;
  editingItem: Record<string, unknown> | null;
}

export function MasterFormFields({ activeTab, editingItem }: MasterFormFieldsProps) {
  // Fetch lookups for forms
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
    enabled: activeTab === "detail-kawasans",
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
                <option key={a.aspek_id} value={a.aspek_id}>{a.aspek_name}</option>
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
                <option key={d.detail_id} value={d.detail_id}>{d.detail_name}</option>
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
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Skor Standar</label>
          <Input
            name="standard_score"
            type="number"
            defaultValue={editingItem?.standard_score as number}
            placeholder="Contoh: 100"
            required
          />
        </div>
      </>
    ),
  };

  return <>{fields[activeTab] || null}</>;
}
