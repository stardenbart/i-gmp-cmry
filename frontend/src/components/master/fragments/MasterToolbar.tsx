import { Search, Plus } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

interface MasterToolbarProps {
  currentTabLabel: string;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  onAddClick: () => void;
}

export function MasterToolbar({
  currentTabLabel,
  searchQuery,
  onSearchChange,
  onAddClick,
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
      <Button onClick={onAddClick}>
        <Plus className="h-4 w-4 mr-2" />
        Tambah {currentTabLabel}
      </Button>
    </div>
  );
}
