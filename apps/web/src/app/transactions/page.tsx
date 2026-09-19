"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, DataTable, EmptyState, ErrorState, LoadingState, StatusPill, type DataTableColumn } from "@nexus/ui";
import { useTransactions } from "@/hooks/use-transactions";
import type { TransactionResponse } from "@/lib/api-types";
import { transactionStatusTone } from "@/lib/status-tone";

const PAGE_LIMIT = 25;

const COLUMNS: readonly DataTableColumn<TransactionResponse>[] = [
  {
    key: "transactionHash",
    header: "Transaction",
    render: (t) => <span className="font-mono text-xs">{t.transactionHash.slice(0, 16)}…</span>,
  },
  { key: "status", header: "Status", render: (t) => <StatusPill tone={transactionStatusTone(t.status)} label={t.status} /> },
  { key: "ledger", header: "Ledger", align: "right", render: (t) => t.ledger ?? "—" },
  { key: "orderId", header: "Order", render: (t) => (t.orderId !== undefined ? `#${t.orderId}` : "—") },
  { key: "lastObservedAt", header: "Last Observed", render: (t) => new Date(t.lastObservedAt).toLocaleString() },
];

export default function TransactionsPage() {
  const router = useRouter();
  const [searchInput, setSearchInput] = useState("");
  const [cursors, setCursors] = useState<readonly string[]>([]);
  const cursor = cursors[cursors.length - 1];

  const query = useTransactions({ ...(cursor !== undefined && { cursor }), limit: PAGE_LIMIT });

  function handleSearch() {
    const hash = searchInput.trim();
    if (hash) {
      router.push(`/transactions/${encodeURIComponent(hash)}`);
    }
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-10">
      <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">Transaction Center</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Every transaction the worker has reconciled against Stellar RPC, and any hash&rsquo;s real current status.
      </p>

      <div className="mt-6 flex gap-2">
        <input
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          placeholder="Look up a transaction hash…"
          className="w-full max-w-md rounded-md border border-slate-300 px-3 py-2 text-sm font-mono dark:border-slate-600 dark:bg-slate-900"
        />
        <Button variant="secondary" onClick={handleSearch} disabled={searchInput.trim() === ""}>
          Look up
        </Button>
      </div>

      <Card className="mt-6">
        {query.isPending && (
          <div className="p-5">
            <LoadingState rows={5} label="Loading transactions…" />
          </div>
        )}
        {query.isError && (
          <div className="p-5">
            <ErrorState message="We could not load recent transactions." detail={query.error.message} onRetry={() => query.refetch()} />
          </div>
        )}
        {query.isSuccess && query.data.transactions.length === 0 && (
          <div className="p-5">
            <EmptyState message="No transactions have been reconciled yet." />
          </div>
        )}
        {query.isSuccess && query.data.transactions.length > 0 && (
          <>
            <DataTable
              columns={COLUMNS}
              rows={query.data.transactions}
              getRowKey={(t) => t.transactionHash}
              onRowClick={(t) => router.push(`/transactions/${encodeURIComponent(t.transactionHash)}`)}
            />
            <div className="flex items-center justify-between border-t border-slate-200 px-5 py-3 dark:border-slate-800">
              <button
                onClick={() => setCursors((prev) => prev.slice(0, -1))}
                disabled={cursors.length === 0}
                className="text-sm font-medium text-slate-600 hover:text-slate-900 disabled:opacity-40 dark:text-slate-400 dark:hover:text-slate-100"
              >
                Previous
              </button>
              <button
                onClick={() => {
                  if (query.data.nextCursor) {
                    setCursors((prev) => [...prev, query.data.nextCursor!]);
                  }
                }}
                disabled={!query.data.nextCursor}
                className="text-sm font-medium text-slate-600 hover:text-slate-900 disabled:opacity-40 dark:text-slate-400 dark:hover:text-slate-100"
              >
                Next
              </button>
            </div>
          </>
        )}
      </Card>
    </div>
  );
}
