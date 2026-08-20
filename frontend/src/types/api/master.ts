/**
 * Master Data API Service
 * Type-safe API calls for master data management
 */

import { api } from "@/lib/api/axios";
import type {
  Department,
  CreateDepartmentRequest,
  UpdateDepartmentRequest,
  Area,
  CreateAreaRequest,
  UpdateAreaRequest,
  Kawasan,
  CreateKawasanRequest,
  UpdateKawasanRequest,
  DetailKawasan,
  CreateDetailKawasanRequest,
  UpdateDetailKawasanRequest,
  Aspek,
  CreateAspekRequest,
  UpdateAspekRequest,
  Detail,
  CreateDetailRequest,
  UpdateDetailRequest,
  Uraian,
  CreateUraianRequest,
  UpdateUraianRequest,
  PaginatedResponse,
  SingleItemResponse,
} from "./index";

/**
 * Generic fetch function for master data
 */
async function fetchMasterData<T>(
  endpoint: string,
  page = 1,
  limit = 10,
  search = "",
  extraParams?: Record<string, string>
): Promise<PaginatedResponse<T>> {
  const res = await api.get(endpoint, {
    params: { page, limit, search, ...extraParams },
  });
  return res.data;
}

async function fetchById<T>(endpoint: string, id: string): Promise<T> {
  const res = await api.get<SingleItemResponse<T>>(`${endpoint}/${id}`);
  return res.data.data;
}

async function createItem<T, R>(
  endpoint: string,
  data: R
): Promise<T> {
  const res = await api.post<SingleItemResponse<T>>(endpoint, data);
  return res.data.data;
}

async function updateItem<T, R>(
  endpoint: string,
  id: string,
  data: R
): Promise<T> {
  const res = await api.put<SingleItemResponse<T>>(`${endpoint}/${id}`, data);
  return res.data.data;
}

async function deleteItem(endpoint: string, id: string): Promise<void> {
  await api.delete(`${endpoint}/${id}`);
}

/**
 * Department API
 */
export const departmentApi = {
  getAll: (
    page = 1,
    limit = 10,
    search = ""
  ): Promise<PaginatedResponse<Department>> =>
    fetchMasterData("/master/departments", page, limit, search),

  getById: (id: string): Promise<Department> =>
    fetchById("/master/departments", id),

  create: (data: CreateDepartmentRequest): Promise<Department> =>
    createItem("/master/departments", data),

  update: (id: string, data: UpdateDepartmentRequest): Promise<Department> =>
    updateItem("/master/departments", id, data),

  delete: (id: string): Promise<void> =>
    deleteItem("/master/departments", id),
};

/**
 * Area API
 */
export const areaApi = {
  getAll: (
    page = 1,
    limit = 10,
    search = "",
    departmentId?: string
  ): Promise<PaginatedResponse<Area>> =>
    fetchMasterData("/master/area", page, limit, search, departmentId ? { department_id: departmentId } : undefined),

  getById: (id: string): Promise<Area> =>
    fetchById("/master/area", id),

  create: (data: CreateAreaRequest): Promise<Area> =>
    createItem("/master/area", data),

  update: (id: string, data: UpdateAreaRequest): Promise<Area> =>
    updateItem("/master/area", id, data),

  delete: (id: string): Promise<void> =>
    deleteItem("/master/area", id),
};

/**
 * Kawasan API
 */
export const kawasanApi = {
  getAll: (
    page = 1,
    limit = 10,
    search = "",
    areaId?: string
  ): Promise<PaginatedResponse<Kawasan>> =>
    fetchMasterData("/master/kawasan", page, limit, search, areaId ? { area_id: areaId } : undefined),

  getById: (id: string): Promise<Kawasan> =>
    fetchById("/master/kawasan", id),

  create: (data: CreateKawasanRequest): Promise<Kawasan> =>
    createItem("/master/kawasan", data),

  update: (id: string, data: UpdateKawasanRequest): Promise<Kawasan> =>
    updateItem("/master/kawasan", id, data),

  delete: (id: string): Promise<void> =>
    deleteItem("/master/kawasan", id),
};

/**
 * Detail Kawasan API
 */
export const detailKawasanApi = {
  getAll: (
    page = 1,
    limit = 10,
    search = "",
    kawasanId?: string
  ): Promise<PaginatedResponse<DetailKawasan>> =>
    fetchMasterData("/master/detail-kawasan", page, limit, search, kawasanId ? { kawasan_id: kawasanId } : undefined),

  getById: (id: string): Promise<DetailKawasan> =>
    fetchById("/master/detail-kawasan", id),

  create: (data: CreateDetailKawasanRequest): Promise<DetailKawasan> =>
    createItem("/master/detail-kawasan", data),

  update: (id: string, data: UpdateDetailKawasanRequest): Promise<DetailKawasan> =>
    updateItem("/master/detail-kawasan", id, data),

  delete: (id: string): Promise<void> =>
    deleteItem("/master/detail-kawasan", id),
};

/**
 * Aspek API
 */
export const aspekApi = {
  getAll: (
    page = 1,
    limit = 10,
    search = "",
    areaId?: string
  ): Promise<PaginatedResponse<Aspek>> =>
    fetchMasterData("/master/aspek", page, limit, search, areaId ? { area_id: areaId } : undefined),

  getById: (id: string): Promise<Aspek> =>
    fetchById("/master/aspek", id),

  create: (data: CreateAspekRequest): Promise<Aspek> =>
    createItem("/master/aspek", data),

  update: (id: string, data: UpdateAspekRequest): Promise<Aspek> =>
    updateItem("/master/aspek", id, data),

  delete: (id: string): Promise<void> =>
    deleteItem("/master/aspek", id),
};

/**
 * Detail API
 */
export const detailApi = {
  getAll: (
    page = 1,
    limit = 10,
    search = "",
    aspekId?: string
  ): Promise<PaginatedResponse<Detail>> =>
    fetchMasterData("/master/details", page, limit, search, aspekId ? { aspek_id: aspekId } : undefined),

  getById: (id: string): Promise<Detail> =>
    fetchById("/master/details", id),

  create: (data: CreateDetailRequest): Promise<Detail> =>
    createItem("/master/details", data),

  update: (id: string, data: UpdateDetailRequest): Promise<Detail> =>
    updateItem("/master/details", id, data),

  delete: (id: string): Promise<void> =>
    deleteItem("/master/details", id),
};

/**
 * Uraian API
 */
export const uraianApi = {
  getAll: (
    page = 1,
    limit = 10,
    search = "",
    detailId?: string
  ): Promise<PaginatedResponse<Uraian>> =>
    fetchMasterData("/master/urain", page, limit, search, detailId ? { detail_id: detailId } : undefined),

  getById: (id: string): Promise<Uraian> =>
    fetchById("/master/urain", id),

  create: (data: CreateUraianRequest): Promise<Uraian> =>
    createItem("/master/urain", data),

  update: (id: string, data: UpdateUraianRequest): Promise<Uraian> =>
    updateItem("/master/urain", id, data),

  delete: (id: string): Promise<void> =>
    deleteItem("/master/urain", id),
};

/**
 * Plant API
 */
export const plantApi = {
  getAll: (
    page = 1,
    limit = 100,
    search = ""
  ): Promise<any> =>
    fetchMasterData("/master/plants", page, limit, search),
};
