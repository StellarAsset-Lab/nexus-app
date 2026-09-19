"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch, toQueryString } from "@/lib/api";
import type { ListEventsResponse, ListTransactionsResponse, OrderSummary, TransactionResponse } from "@/lib/api-types";

export interface UseTransactionsFilter {
  readonly status?: string;
  readonly cursor?: string;
  readonly limit?: number;
}

export function useTransactions(filter: UseTransactionsFilter = {}) {
  return useQuery({
    queryKey: ["transactions", filter],
    queryFn: () =>
      apiFetch<ListTransactionsResponse>(
        `/api/v1/transactions${toQueryString({ status: filter.status, cursor: filter.cursor, limit: filter.limit })}`,
      ),
  });
}

export function useTransaction(hash: string | undefined) {
  return useQuery({
    queryKey: ["transaction", hash],
    queryFn: () => apiFetch<TransactionResponse>(`/api/v1/transactions/${encodeURIComponent(hash!)}`),
    enabled: hash !== undefined,
    retry: false,
  });
}

export function useTransactionEvents(hash: string | undefined) {
  return useQuery({
    queryKey: ["transaction-events", hash],
    queryFn: () =>
      apiFetch<ListEventsResponse>(`/api/v1/events${toQueryString({ transactionHash: hash, limit: 200 })}`),
    enabled: hash !== undefined,
  });
}

export function useOrder(orderId: number | undefined) {
  return useQuery({
    queryKey: ["order", orderId],
    queryFn: () => apiFetch<OrderSummary>(`/api/v1/orders/${orderId}`),
    enabled: orderId !== undefined,
  });
}
