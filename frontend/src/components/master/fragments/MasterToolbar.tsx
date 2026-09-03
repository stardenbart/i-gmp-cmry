import { Search, Plus, Upload } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface MasterToolbarProps {
  currentTabLabel: string;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  onAddClick: () => void;
  canImport?: boolean;
  onImportClick?: () => void;
}

export function MasterToolbar({
  currentTabLabel,
  searchQuery,
  onSearchChange,
  onAddClick,
  canImport = false,
  onImportClick,
}: MasterToolbarProps) {
  return (
    <div className="flex flex-col sm:flex-row gap-4 justify-between">
      <div className="relative w-full sm:max-w-sm min-w-[250px]">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
        <Input
          placeholder={`Cari ${currentTabLabel.toLowerCase()}...`}
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-9 w-full"
        />
      </div>
      <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
        {canImport && onImportClick && (
          <Button variant="outline" onClick={onImportClick} className="w-full sm:w-auto">
            <Upload className="h-4 w-4" />
            Import Excel
          </Button>
        )}
        <Button onClick={onAddClick} className="w-full sm:w-auto">
          <Plus className="h-4 w-4" />
          Tambah {currentTabLabel}
        </Button>
      </div>
    </div>
  );
}
