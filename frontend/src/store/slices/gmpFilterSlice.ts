import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface GmpFilterState {
  selectedArea: string;
  selectedKawasan: string;
  selectedDetailKawasan: string;
  startDate: string;
  endDate: string;
  q: string;
}

const initialState: GmpFilterState = {
  selectedArea: "",
  selectedKawasan: "",
  selectedDetailKawasan: "",
  startDate: "",
  endDate: "",
  q: "",
};

export const gmpFilterSlice = createSlice({
  name: "gmpFilter",
  initialState,
  reducers: {
    setSelectedArea: (state, action: PayloadAction<string>) => {
      state.selectedArea = action.payload;
      state.selectedKawasan = "";
      state.selectedDetailKawasan = "";
    },
    setSelectedKawasan: (state, action: PayloadAction<string>) => {
      state.selectedKawasan = action.payload;
      state.selectedDetailKawasan = "";
    },
    setSelectedDetailKawasan: (state, action: PayloadAction<string>) => {
      state.selectedDetailKawasan = action.payload;
    },
    setStartDate: (state, action: PayloadAction<string>) => {
      state.startDate = action.payload;
    },
    setEndDate: (state, action: PayloadAction<string>) => {
      state.endDate = action.payload;
    },
    setQ: (state, action: PayloadAction<string>) => {
      state.q = action.payload;
    },
    resetFilters: (state) => {
      state.selectedArea = "";
      state.selectedKawasan = "";
      state.selectedDetailKawasan = "";
      state.startDate = "";
      state.endDate = "";
      state.q = "";
    },
  },
});

export const {
  setSelectedArea,
  setSelectedKawasan,
  setSelectedDetailKawasan,
  setStartDate,
  setEndDate,
  setQ,
  resetFilters,
} = gmpFilterSlice.actions;

export default gmpFilterSlice.reducer;
