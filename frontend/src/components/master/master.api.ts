import { api } from "@/lib/api/axios";

export const fetchItems = async (endpoint: string, page = 1, search = "", limit = 10) => {
  const res = await api.get(endpoint, { params: { page, limit, search } });
  const rawData = res.data;
  const responseData = rawData.data;

  const items = Array.isArray(responseData)
    ? responseData
    : responseData?.items || [];

  const total = Array.isArray(responseData)
    ? rawData.meta?.total || items.length
    : responseData?.total || rawData.meta?.total || 0;

  const totalPages = Math.ceil(total / limit) || 1;

  return {
    items,
    pagination: {
      total,
      page: rawData.meta?.page || responseData?.page || page,
      limit: rawData.meta?.limit || responseData?.limit || limit,
      total_pages: rawData.meta?.total_pages || responseData?.total_pages || totalPages,
    },
  };
};

export const createItem = async (endpoint: string, data: Record<string, unknown>) => {
  const res = await api.post(endpoint, data);
  return res.data;
};

export const updateItem = async (endpoint: string, id: string, data: Record<string, unknown>) => {
  const res = await api.put(`${endpoint}/${id}`, data);
  return res.data;
};

export const deleteItem = async (endpoint: string, id: string) => {
  const res = await api.delete(`${endpoint}/${id}`);
  return res.data;
};
