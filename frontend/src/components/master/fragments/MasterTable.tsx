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
                      {col.key === "aspek_weight"
                        ? `${item[col.key]}%`
                        : ((item[col.key] as string) ||
                           (col.key === "department_name" ? (item.department as any)?.department_name : "") ||
                           (col.key === "area_name" ? (item.area as any)?.area_name : "") ||
                           (col.key === "kawasan_name" ? (item.kawasan as any)?.kawasan_name : "") ||
                           "-")}
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
