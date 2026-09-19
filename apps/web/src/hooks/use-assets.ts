"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch, toQueryString } from "@/lib/api";
import type { AssetActivityResponse, AssetResponse, ListAssetsResponse } from "@/lib/api-types";

export interface UseAssetsFilter {
  readonly active?: boolean;
  readonly cursor?: string;
  readonly limit?: number;
}

export function useAssets(filter: UseAssetsFilter = {}) {
  return useQuery({
    queryKey: ["assets", filter],
    queryFn: () =>
      apiFetch<ListAssetsResponse>(
        `/api/v1/assets${toQueryString({ active: filter.active, cursor: filter.cursor, limit: filter.limit })}`,
      ),
  });
}

export function useAsset(asset: string | undefined) {
  return useQuery({
    queryKey: ["asset", asset],
    queryFn: () => apiFetch<AssetResponse>(`/api/v1/assets/${encodeURIComponent(asset!)}`),
    enabled: asset !== undefined,
  });
}

export function useAssetActivity(asset: string | undefined) {
  return useQuery({
    queryKey: ["asset-activity", asset],
    queryFn: () => apiFetch<AssetActivityResponse>(`/api/v1/assets/${encodeURIComponent(asset!)}/activity`),
    enabled: asset !== undefined,
  });
}
