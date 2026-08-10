import { useQuery } from "@tanstack/react-query";
import { fetchItems } from "@/components/master/master.api";

const MASTER_STALE_TIME = 10 * 60 * 1000; // 10 minutes cache freshness
const MASTER_GC_TIME = 30 * 60 * 1000;    // 30 minutes memory persistence

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
