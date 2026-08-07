import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface SettingsState {
  activeTab: "general" | "email" | "apikey";
  plantFilter: string;
}

const initialState: SettingsState = {
  activeTab: "general",
  plantFilter: "ALL",
};

export const settingsSlice = createSlice({
  name: "settings",
  initialState,
  reducers: {
    setActiveTab: (state, action: PayloadAction<"general" | "email" | "apikey">) => {
      state.activeTab = action.payload;
    },
    setPlantFilter: (state, action: PayloadAction<string>) => {
      state.plantFilter = action.payload;
    },
  },
});

export const { setActiveTab, setPlantFilter } = settingsSlice.actions;

export default settingsSlice.reducer;
