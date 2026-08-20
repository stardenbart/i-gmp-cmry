import { X } from "lucide-react";
import { Button } from "@/components/ui/button";

export function MasterFormModal({
  isOpen,
  onClose,
  title,
  onSubmit,
  isLoading,
  children,
}: {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  isLoading: boolean;
  children: React.ReactNode;
}) {
  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm transition-opacity"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="w-[90vw] sm:w-[450px] bg-card rounded-2xl shadow-2xl border border-border overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <form onSubmit={onSubmit}>
          <div className="px-6 py-4 border-b border-border/50 flex items-center justify-between bg-card">
            <h2 className="text-lg font-semibold text-foreground">{title}</h2>
            <button type="button" onClick={onClose} className="p-1.5 hover:bg-muted rounded-lg text-muted-foreground hover:text-foreground transition-colors">
              <X className="h-5 w-5" />
            </button>
          </div>
          
          <div className="p-6 max-h-[60vh] overflow-y-auto space-y-4 bg-card">
            {children}
          </div>
          
          <div className="px-6 py-4 border-t border-border/50 flex gap-3 bg-muted/10">
            <Button type="button" variant="outline" onClick={onClose} className="flex-1">
              Batal
            </Button>
            <Button type="submit" isLoading={isLoading} className="flex-1">
              Simpan
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

export function MasterDeleteModal({
  isOpen,
  onClose,
  onConfirm,
  isLoading,
  itemName,
}: {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isLoading: boolean;
  itemName: string;
}) {
  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm transition-opacity"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="w-[90vw] sm:w-[400px] bg-card rounded-2xl shadow-2xl border border-border overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div className="p-6">
          <h2 className="text-lg font-semibold mb-2 text-foreground">Hapus Data</h2>
          <p className="text-sm text-muted-foreground mb-6">
            Apakah Anda yakin ingin menghapus <strong>{itemName}</strong>? Tindakan ini tidak dapat dibatalkan.
          </p>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={onClose} className="flex-1">
              Batal
            </Button>
            <Button variant="destructive" onClick={onConfirm} isLoading={isLoading} className="flex-1">
              Hapus
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
