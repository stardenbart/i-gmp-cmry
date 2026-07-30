import { Settings, Edit2, Trash2, ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export interface ColumnDef {
  key: string;
  label: string;
}

interface MasterTableProps {
  items: Record<string, unknown>[];
  columns: ColumnDef[];
  isLoading: boolean;
  currentTabLabel: string;
  idField: string;
  nameField: string;
  pagination: {
    total: number;
    page: number;
    limit: number;
    total_pages: number;
  };
  onPageChange: (page: number) => void;
  onEdit: (item: Record<string, unknown>) => void;
  onDelete: (id: string, name: string) => void;
}

export function MasterTable({
  items,
  columns,
  isLoading,
  currentTabLabel,
  idField,
  nameField,
  pagination,
  onPageChange,
  onEdit,
  onDelete,
}: MasterTableProps) {
  return (
    <div className="rounded-xl border border-border bg-card overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              {columns.map((col) => (
                <th
                  key={col.key}
                  className="px-4 py-3 text-left text-sm font-medium text-muted-foreground whitespace-nowrap"
                >
                  {col.label}
                </th>
              ))}
              <th className="px-4 py-3 text-right text-sm font-medium text-muted-foreground whitespace-nowrap">
                Aksi
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {isLoading ? (
              <tr>
                <td colSpan={columns.length + 1}>
                  <div className="flex justify-center py-12">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  </div>
                </td>
              </tr>
            ) : items.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length + 1}
                  className="px-4 py-12 text-center text-muted-foreground"
                >
                  <Settings className="h-12 w-12 mx-auto mb-3 opacity-20" />
                  <p>Belum ada data {currentTabLabel.toLowerCase()}</p>
                  <p className="text-sm">Klik tombol tambah untuk membuat data baru</p>
                </td>
              </tr>
            ) : (
              items.map((item) => (
                <tr key={item[idField] as string} className="hover:bg-muted/50 transition-colors">
                  {columns.map((col) => (
                    <td key={col.key} className="px-4 py-3 text-sm">
                      {(() => {
                        const raw = item[col.key];
                        let val = "";

                        if (raw !== undefined && raw !== null && raw !== "") {
                          val = String(raw);
                        } else {
                          // Lookups for 1-level and 2-level parent relations
                          const lookupMap: Record<string, () => string | undefined> = {
                            "department_name":   () => (item.department as any)?.department_name,
                            "area_name":         () => (item.area as any)?.area_name || (item.kawasan as any)?.area?.area_name || (item.aspek as any)?.area?.area_name,
                            "kawasan_name":      () => (item.kawasan as any)?.kawasan_name,
                            "aspek_name":        () => (item.aspek as any)?.aspek_name,
                            "detail_name":       () => (item.detail as any)?.detail_name,
                            "kawasan_area_name": () => (item.kawasan as any)?.area?.area_name,
                            "aspek_area_name":   () => (item.aspek as any)?.area?.area_name,
                            "detail_aspek_name": () => (item.detail as any)?.aspek?.aspek_name,
                          };

                          val = lookupMap[col.key]?.() || "-";
                        }

                        if (col.key === "aspek_weight") {
                          val = `${val}%`;
                        }

                        const isRelationCol = [
                          "area_name",
                          "kawasan_name",
                          "aspek_name",
                          "detail_name",
                          "kawasan_area_name",
                          "aspek_area_name",
                          "detail_aspek_name",
                        ].includes(col.key);

                        if (isRelationCol && val !== "-") {
                          return (
                            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs bg-secondary text-secondary-foreground font-medium border border-border/50">
                              {val}
                            </span>
                          );
                        }

                        return val;
                      })()}
                    </td>
                  ))}
                  <td className="px-4 py-3 text-right">
                    <div className="flex justify-end gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onEdit(item)}
                        className="h-8 w-8 p-0"
                      >
                        <Edit2 className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          onDelete(item[idField] as string, item[nameField] as string)
                        }
                        className="h-8 w-8 p-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {pagination.total_pages > 1 && (
        <div className="flex items-center justify-between px-4 py-3 border-t border-border bg-muted/20">
          <p className="text-sm text-muted-foreground">
            Menampilkan {(pagination.page - 1) * pagination.limit + 1} hingga{" "}
            {Math.min(pagination.page * pagination.limit, pagination.total)} dari{" "}
            {pagination.total} data
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(Math.max(1, pagination.page - 1))}
              disabled={pagination.page === 1}
            >
              <ChevronLeft className="h-4 w-4 mr-1" />
              Sebelumnya
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(Math.min(pagination.total_pages, pagination.page + 1))}
              disabled={pagination.page === pagination.total_pages}
            >
              Selanjutnya
              <ChevronRight className="h-4 w-4 ml-1" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
