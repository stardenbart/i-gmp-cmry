import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export interface UserFilterState {
  searchQuery: string;
  roleFilter: string;
  plantFilter: string;
  statusFilter: string;
  page: number;
}

const initialState: UserFilterState = {
  searchQuery: "",
  roleFilter: "ALL",
  plantFilter: "ALL",
  statusFilter: "ALL",
  page: 1,
};

export const userFilterSlice = createSlice({
  name: "userFilter",
  initialState,
  reducers: {
    setSearchQuery: (state, action: PayloadAction<string>) => {
      state.searchQuery = action.payload;
      state.page = 1;
    },
    setRoleFilter: (state, action: PayloadAction<string>) => {
      state.roleFilter = action.payload;
      state.page = 1;
    },
    setPlantFilter: (state, action: PayloadAction<string>) => {
      state.plantFilter = action.payload;
      state.page = 1;
    },
    setStatusFilter: (state, action: PayloadAction<string>) => {
      state.statusFilter = action.payload;
      state.page = 1;
    },
    setPage: (state, action: PayloadAction<number>) => {
      state.page = action.payload;
    },
    resetFilters: (state) => {
      state.searchQuery = "";
      state.roleFilter = "ALL";
      state.plantFilter = "ALL";
      state.statusFilter = "ALL";
      state.page = 1;
    },
  },
});

export const {
  setSearchQuery,
  setRoleFilter,
  setPlantFilter,
  setStatusFilter,
  setPage,
  resetFilters,
} = userFilterSlice.actions;

export default userFilterSlice.reducer;
