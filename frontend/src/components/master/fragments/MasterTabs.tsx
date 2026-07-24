import { MASTER_TABS } from "../master.types";

interface MasterTabsProps {
  activeTab: string;
  onTabChange: (tabId: string) => void;
}

export function MasterTabs({ activeTab, onTabChange }: MasterTabsProps) {
  return (
    <div className="border-b border-border">
      <div className="flex gap-1 overflow-x-auto pb-px -mb-px scrollbar-hide">
        {MASTER_TABS.map((tab) => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => onTabChange(tab.id)}
              className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium whitespace-nowrap border-b-2 transition-colors ${
                activeTab === tab.id
                  ? "border-primary text-primary"
                  : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
              }`}
            >
              <Icon className="h-4 w-4" />
              {tab.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}
