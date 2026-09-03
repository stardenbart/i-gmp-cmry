import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface IssueFilterState {
  activeStatus: string;
  search: string;
  dateFrom: string;
  dateTo: string;
  page: number;
}

const initialState: IssueFilterState = {
  activeStatus: "all",
  search: "",
  dateFrom: "",
  dateTo: "",
  page: 1,
};

export const issueFilterSlice = createSlice({
  name: "issueFilter",
  initialState,
  reducers: {
    setActiveStatus: (state, action: PayloadAction<string>) => {
      state.activeStatus = action.payload;
      state.page = 1;
    },
    setSearch: (state, action: PayloadAction<string>) => {
      state.search = action.payload;
      state.page = 1;
    },
    setDateFrom: (state, action: PayloadAction<string>) => {
      state.dateFrom = action.payload;
      state.page = 1;
    },
    setDateTo: (state, action: PayloadAction<string>) => {
      state.dateTo = action.payload;
      state.page = 1;
    },
    setPage: (state, action: PayloadAction<number>) => {
      state.page = action.payload;
    },
    resetFilters: (state) => {
      state.activeStatus = "all";
      state.search = "";
      state.dateFrom = "";
      state.dateTo = "";
      state.page = 1;
    },
  },
});

export const {
  setActiveStatus,
  setSearch,
  setDateFrom,
  setDateTo,
  setPage,
  resetFilters,
} = issueFilterSlice.actions;

export default issueFilterSlice.reducer;
