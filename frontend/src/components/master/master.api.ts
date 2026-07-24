import { api } from "@/lib/api/axios";

export const fetchItems = async (endpoint: string, page = 1, search = "", limit = 10) => {
  const res = await api.get(endpoint, { params: { page, limit, search } });
  const responseData = res.data.data;
  return {
    items: responseData.items || [],
    pagination: {
      total: responseData.total || 0,
      page: responseData.page || 1,
      limit: responseData.limit || limit,
      total_pages: responseData.total_pages || 1,
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
