"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api";
import type { NetworkResponse, StatusResponse } from "@/lib/api-types";

export function useStatus() {
  return useQuery({
    queryKey: ["status"],
    queryFn: () => apiFetch<StatusResponse>("/api/v1/status"),
    // Status should reflect near-real-time reality, not the default 15s
    // staleness other indexed-data views use.
    staleTime: 5_000,
    refetchInterval: 15_000,
  });
}

export function useNetwork() {
  return useQuery({
    queryKey: ["network"],
    queryFn: () => apiFetch<NetworkResponse>("/api/v1/network"),
    staleTime: 5_000,
    refetchInterval: 15_000,
  });
}
