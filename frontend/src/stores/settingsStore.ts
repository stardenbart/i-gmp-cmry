import { create } from "zustand";
import { persist } from "zustand/middleware";

interface SettingsState {
  showSearchLatencyButton: boolean;
  setShowSearchLatencyButton: (show: boolean) => void;
  toggleSearchLatencyButton: () => void;
  showCoreWebVitalsMonitor: boolean;
  setShowCoreWebVitalsMonitor: (show: boolean) => void;
  toggleCoreWebVitalsMonitor: () => void;
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      showSearchLatencyButton: true,
      setShowSearchLatencyButton: (show: boolean) => set({ showSearchLatencyButton: show }),
      toggleSearchLatencyButton: () => set((state) => ({ showSearchLatencyButton: !state.showSearchLatencyButton })),
      showCoreWebVitalsMonitor: false, // Default to FALSE so it is 100% hidden unless Admin explicitly enables it
      setShowCoreWebVitalsMonitor: (show: boolean) => set({ showCoreWebVitalsMonitor: show }),
      toggleCoreWebVitalsMonitor: () => set((state) => ({ showCoreWebVitalsMonitor: !state.showCoreWebVitalsMonitor })),
    }),
    {
      name: "app-settings-storage",
    }
  )
);
