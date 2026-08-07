import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface LogFilterState {
  activeTab: "activity" | "login";
  page: number;
  searchQuery: string;
  userFilter: string;
  plantFilter: string;
}

const initialState: LogFilterState = {
  activeTab: "activity",
  page: 1,
  searchQuery: "",
  userFilter: "ALL",
  plantFilter: "ALL",
};

export const logFilterSlice = createSlice({
  name: "logFilter",
  initialState,
  reducers: {
    setActiveTab: (state, action: PayloadAction<"activity" | "login">) => {
      state.activeTab = action.payload;
      state.page = 1;
    },
    setPage: (state, action: PayloadAction<number>) => {
      state.page = action.payload;
    },
    setSearchQuery: (state, action: PayloadAction<string>) => {
      state.searchQuery = action.payload;
      state.page = 1;
    },
    setUserFilter: (state, action: PayloadAction<string>) => {
      state.userFilter = action.payload;
      state.page = 1;
    },
    setPlantFilter: (state, action: PayloadAction<string>) => {
      state.plantFilter = action.payload;
      state.page = 1;
    },
    resetFilters: (state) => {
      state.activeTab = "activity";
      state.page = 1;
      state.searchQuery = "";
      state.userFilter = "ALL";
      state.plantFilter = "ALL";
    },
  },
});

export const {
  setActiveTab,
  setPage,
  setSearchQuery,
  setUserFilter,
  setPlantFilter,
  resetFilters,
} = logFilterSlice.actions;

export default logFilterSlice.reducer;
