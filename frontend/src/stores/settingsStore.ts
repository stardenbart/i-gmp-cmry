import { create } from "zustand";
import { persist } from "zustand/middleware";

interface SettingsState {
  showSearchLatencyButton: boolean;
  setShowSearchLatencyButton: (show: boolean) => void;
  toggleSearchLatencyButton: () => void;
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      showSearchLatencyButton: true,
      setShowSearchLatencyButton: (show: boolean) => set({ showSearchLatencyButton: show }),
      toggleSearchLatencyButton: () => set((state) => ({ showSearchLatencyButton: !state.showSearchLatencyButton })),
    }),
    {
      name: "app-settings-storage",
    }
  )
);
