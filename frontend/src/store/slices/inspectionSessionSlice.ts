import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { PhotoItem } from "@/components/Inspection/PhotoUploaderWithKeterangan";

export interface InspectionSessionState {
  activeAspekIndex: number;
  activeDetailIndex: number | "all";
  highlightedUraianId: string | null;
  ngPhotosMap: Record<string, PhotoItem[]>;
  showCancelDialog: boolean;
  showScrollTop: boolean;
}

const initialState: InspectionSessionState = {
  activeAspekIndex: 0,
  activeDetailIndex: 0,
  highlightedUraianId: null,
  ngPhotosMap: {},
  showCancelDialog: false,
  showScrollTop: false,
};

export const inspectionSessionSlice = createSlice({
  name: "inspectionSession",
  initialState,
  reducers: {
    setActiveAspekIndex: (state, action: PayloadAction<number>) => {
      state.activeAspekIndex = action.payload;
    },
    setActiveDetailIndex: (state, action: PayloadAction<number | "all">) => {
      state.activeDetailIndex = action.payload;
    },
    setHighlightedUraianId: (state, action: PayloadAction<string | null>) => {
      state.highlightedUraianId = action.payload;
    },
    setNgPhotosMap: (state, action: PayloadAction<Record<string, PhotoItem[]>>) => {
      state.ngPhotosMap = action.payload;
    },
    updateItemPhotos: (
      state,
      action: PayloadAction<{ pKey: string; photos: PhotoItem[] }>
    ) => {
      state.ngPhotosMap[action.payload.pKey] = action.payload.photos;
    },
    setShowCancelDialog: (state, action: PayloadAction<boolean>) => {
      state.showCancelDialog = action.payload;
    },
    setShowScrollTop: (state, action: PayloadAction<boolean>) => {
      state.showScrollTop = action.payload;
    },
    resetSession: (state) => {
      state.activeAspekIndex = 0;
      state.activeDetailIndex = 0;
      state.highlightedUraianId = null;
      state.ngPhotosMap = {};
      state.showCancelDialog = false;
      state.showScrollTop = false;
    },
  },
});

export const {
  setActiveAspekIndex,
  setActiveDetailIndex,
  setHighlightedUraianId,
  setNgPhotosMap,
  updateItemPhotos,
  setShowCancelDialog,
  setShowScrollTop,
  resetSession,
} = inspectionSessionSlice.actions;

export default inspectionSessionSlice.reducer;
