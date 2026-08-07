import { configureStore } from "@reduxjs/toolkit";
import inspectionSessionReducer from "./slices/inspectionSessionSlice";
import inspectionFilterReducer from "./slices/inspectionFilterSlice";
import userFilterReducer from "./slices/userFilterSlice";
import issueFilterReducer from "./slices/issueFilterSlice";
import logFilterReducer from "./slices/logFilterSlice";
import gmpFilterReducer from "./slices/gmpFilterSlice";
import settingsReducer from "./slices/settingsSlice";

export const store = configureStore({
  reducer: {
    inspectionSession: inspectionSessionReducer,
    inspectionFilter: inspectionFilterReducer,
    userFilter: userFilterReducer,
    issueFilter: issueFilterReducer,
    logFilter: logFilterReducer,
    gmpFilter: gmpFilterReducer,
    settings: settingsReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false,
      immutableCheck: false,
    }),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
