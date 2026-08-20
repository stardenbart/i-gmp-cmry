import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface InspectionFilterState {
  q: string;
  status: string;
  page: number;
  limit: number;
}

const initialState: InspectionFilterState = {
  q: "",
  status: "",
  page: 1,
  limit: 20,
};

export const inspectionFilterSlice = createSlice({
  name: "inspectionFilter",
  initialState,
  reducers: {
    setSearchQuery: (state, action: PayloadAction<string>) => {
      state.q = action.payload;
      state.page = 1; // Reset to page 1 on search change
    },
    setStatusFilter: (state, action: PayloadAction<string>) => {
      state.status = action.payload;
      state.page = 1; // Reset to page 1 on filter change
    },
    setPage: (state, action: PayloadAction<number>) => {
      state.page = action.payload;
    },
    resetFilters: (state) => {
      state.q = "";
      state.status = "";
      state.page = 1;
    },
  },
});

export const {
  setSearchQuery,
  setStatusFilter,
  setPage,
  resetFilters,
} = inspectionFilterSlice.actions;

export default inspectionFilterSlice.reducer;
