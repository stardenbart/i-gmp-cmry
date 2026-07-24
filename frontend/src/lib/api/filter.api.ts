import { api } from "./axios";

// ─── Shared Filter Types ─────────────────────────────────────────────────

export interface DateRangeFacet {
  min: string | null;
  max: string | null;
  field: string;
}

export interface FilterResult<T, F> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
  filters_applied: Record<string, any>;
  facets: F;
}

// ─── Inspection Filter Types ──────────────────────────────────────────────

export interface InspectionFacets {
  status: Record<string, number>;
  area_id: Record<string, number>;
  kawasan_id: Record<string, number>;
  inspector_id: Record<string, number>;
  date_range: DateRangeFacet;
}

export interface InspectionFilterParams {
  q?: string;
  status?: string;
  status__in?: string;
  area_id?: string;
  area_id__in?: string;
  kawasan_id?: string;
  kawasan_id__in?: string;
  detail_kawasan_id?: string;
  inspector_id?: string;
  inspector_id__in?: string;
  date_from?: string;
  date_to?: string;
  sort_by?: string;
  sort_order?: string;
  page?: number;
  limit?: number;
}

// ─── Issue Filter Types ───────────────────────────────────────────────────

export interface IssueFacets {
  status: Record<string, number>;
  wowr_status: Record<string, number>;
  issue_pic_user_id: Record<string, number>;
  date_range: DateRangeFacet;
  due_date_range: DateRangeFacet;
}

export interface IssueFilterParams {
  q?: string;
  status?: string;
  status__in?: string;
  issue_pic_user_id?: string;
  wowr_status?: string;
  wowr_status__in?: string;
  needs_wo_wr?: string;
  label?: string;
  date_from?: string;
  date_to?: string;
  due_from?: string;
  due_to?: string;
  sort_by?: string;
  sort_order?: string;
  page?: number;
  limit?: number;
}

// ─── Followup Filter Types ────────────────────────────────────────────────

export interface FollowupFacets {
  status: Record<string, number>;
  wowr_status: Record<string, number>;
  date_range: DateRangeFacet;
  due_date_range: DateRangeFacet;
}

export interface FollowupFilterParams extends IssueFilterParams {
  include_closed?: string;
}

// ─── User Filter Types ────────────────────────────────────────────────────

export interface UserFacets {
  role_id: Record<string, number>;
  department_id: Record<string, number>;
  user_status: Record<string, number>;
  date_range: DateRangeFacet;
}

export interface UserFilterParams {
  q?: string;
  role_id?: string;
  role_id__in?: string;
  department_id?: string;
  user_status?: string;
  user_status__in?: string;
  date_from?: string;
  date_to?: string;
  sort_by?: string;
  sort_order?: string;
  page?: number;
  limit?: number;
}

// ─── API Functions ────────────────────────────────────────────────────────

export const filterApi = {
  inspections: async (params?: InspectionFilterParams) => {
    const res = await api.get("/inspections/filter", { params });
    return res.data.data as FilterResult<any, InspectionFacets>;
  },

  issues: async (params?: IssueFilterParams) => {
    const res = await api.get("/issues/filter", { params });
    return res.data.data as FilterResult<any, IssueFacets>;
  },

  followup: async (params?: FollowupFilterParams) => {
    const res = await api.get("/issues/followup/filter", { params });
    return res.data.data as FilterResult<any, FollowupFacets>;
  },

  users: async (params?: UserFilterParams) => {
    const res = await api.get("/users/filter", { params });
    return res.data.data as FilterResult<any, UserFacets>;
  },
};
