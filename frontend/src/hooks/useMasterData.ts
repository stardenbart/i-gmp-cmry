import { useQuery } from "@tanstack/react-query";
import { fetchItems } from "@/components/master/master.api";
import { api } from "@/lib/api/axios";

export interface HEIMasterItem {
  hei_id: string;
  category_name: string;
  item_name?: string;
  hei_name?: string;
  hei_code?: string;
}

const MASTER_STALE_TIME = 10 * 60 * 1000; // 10 minutes cache freshness
const MASTER_GC_TIME = 30 * 60 * 1000;    // 30 minutes memory persistence

export function useMasterHEI(category = "", limit = 500) {
  return useQuery({
    queryKey: ["master", "hei", category, limit],
    queryFn: () => {
      const url = category ? `/master/hei?category=${encodeURIComponent(category)}` : "/master/hei";
      return fetchItems<HEIMasterItem>(url, 1, "", limit);
    },
    staleTime: MASTER_STALE_TIME,
    gcTime: MASTER_GC_TIME,
  });
}

export function useMasterHEICategories() {
  return useQuery({
    queryKey: ["master", "hei", "categories"],
    queryFn: async () => {
      const res = await api.get<{ data: string[] }>("/master/hei/categories");
      return res.data.data || ["Habit", "Equipment", "Infrastructure"];
    },
    staleTime: MASTER_STALE_TIME,
    gcTime: MASTER_GC_TIME,
  });
}

export function useMasterHabits(limit = 500) {
  return useQuery({
    queryKey: ["master", "habits", limit],
    queryFn: () => fetchItems("/master/habits", 1, "", limit),
    staleTime: MASTER_STALE_TIME,
    gcTime: MASTER_GC_TIME,
  });
}

export function useMasterEquipments(limit = 500) {
  return useQuery({
    queryKey: ["master", "equipments", limit],
    queryFn: () => fetchItems("/master/equipments", 1, "", limit),
    staleTime: MASTER_STALE_TIME,
    gcTime: MASTER_GC_TIME,
  });
}

export function useMasterInfrastructures(limit = 500) {
  return useQuery({
    queryKey: ["master", "infrastructures", limit],
    queryFn: () => fetchItems("/master/infrastructures", 1, "", limit),
    staleTime: MASTER_STALE_TIME,
    gcTime: MASTER_GC_TIME,
  });
}

export function useMasterAreas(limit = 500) {
  return useQuery({
    queryKey: ["master", "areas", limit],
    queryFn: () => fetchItems("/master/areas", 1, "", limit),
    staleTime: MASTER_STALE_TIME,
    gcTime: MASTER_GC_TIME,
  });
}

export function useMasterPlants(limit = 500) {
  return useQuery({
    queryKey: ["master", "plants", limit],
    queryFn: () => fetchItems("/master/plants", 1, "", limit),
    staleTime: MASTER_STALE_TIME,
    gcTime: MASTER_GC_TIME,
  });
}
