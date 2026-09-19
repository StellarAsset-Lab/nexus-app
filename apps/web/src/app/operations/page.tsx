"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Badge, Card, CardBody, CardHeader, DataTable, EmptyState, ErrorState, LoadingState, StatusPill, type DataTableColumn } from "@nexus/ui";
import { useNetwork } from "@/hooks/use-status";
import { useTransactions } from "@/hooks/use-transactions";
import type { TransactionResponse } from "@/lib/api-types";
import { transactionStatusTone } from "@/lib/status-tone";

// A "Pending" transaction the worker has not yet reconciled to a terminal
// status within this window is surfaced as stale — long enough past a
// normal confirmation to be worth operator attention, per
// internal/reconcile's own indexerLagThreshold-style reasoning.
const STALE_THRESHOLD_MS = 5 * 60 * 1000;

function staleColumns(now: number): readonly DataTableColumn<TransactionResponse>[] {
  return [
    {
      key: "transactionHash",
      header: "Transaction",
      render: (t) => (
        <Link href={`/transactions/${encodeURIComponent(t.transactionHash)}`} className="font-mono text-xs text-blue-600 hover:underline dark:text-blue-400">
          {t.transactionHash.slice(0, 16)}…
        </Link>
      ),
    },
    { key: "status", header: "Status", render: (t) => <StatusPill tone={transactionStatusTone(t.status)} label={t.status} /> },
    { key: "firstObservedAt", header: "First Observed", render: (t) => new Date(t.firstObservedAt).toLocaleString() },
    {
      key: "age",
      header: "Age",
      align: "right",
      render: (t) => formatAge(now - new Date(t.firstObservedAt).getTime()),
    },
  ];
}

function formatAge(ms: number): string {
  const minutes = Math.floor(ms / 60_000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ${minutes % 60}m`;
}

export default function OperationsPage() {
  const networkQuery = useNetwork();
  const pendingQuery = useTransactions({ status: "Pending", limit: 100 });

  // `now` is read from state, refreshed on an interval, rather than calling
  // Date.now() directly during render — components must stay pure per the
  // React Compiler's rules.
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 30_000);
    return () => clearInterval(id);
  }, []);

  const ingestLag =
    networkQuery.data?.latestLedger !== undefined && networkQuery.data.indexerCheckpoint !== undefined
      ? networkQuery.data.latestLedger - networkQuery.data.indexerCheckpoint
      : undefined;

  const staleTransactions = (pendingQuery.data?.transactions ?? []).filter(
    (t) => now - new Date(t.firstObservedAt).getTime() > STALE_THRESHOLD_MS,
  );

  return (
    <div className="mx-auto max-w-7xl px-4 py-10">
      <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">Operations Console</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Indexer checkpoint, ingest lag, and stale transactions the worker has not yet reconciled — operator-facing,
        never customer-facing.
      </p>

      <div className="mt-6 grid gap-4 sm:grid-cols-3">
        <Card>
          <CardBody>
            <p className="text-xs uppercase tracking-wide text-slate-500 dark:text-slate-400">Indexer checkpoint</p>
            {networkQuery.isPending && <LoadingState rows={1} label="Loading…" />}
            {networkQuery.isSuccess && (
              <p className="mt-1 text-2xl font-semibold text-slate-900 dark:text-slate-100">
                {networkQuery.data.indexerCheckpoint ?? "None yet"}
              </p>
            )}
          </CardBody>
        </Card>
        <Card>
          <CardBody>
            <p className="text-xs uppercase tracking-wide text-slate-500 dark:text-slate-400">Latest ledger</p>
            {networkQuery.isPending && <LoadingState rows={1} label="Loading…" />}
            {networkQuery.isSuccess && (
              <p className="mt-1 text-2xl font-semibold text-slate-900 dark:text-slate-100">
                {networkQuery.data.latestLedger ?? "Unknown"}
              </p>
            )}
          </CardBody>
        </Card>
        <Card>
          <CardBody>
            <p className="text-xs uppercase tracking-wide text-slate-500 dark:text-slate-400">Ingest lag</p>
            {networkQuery.isPending && <LoadingState rows={1} label="Loading…" />}
            {networkQuery.isSuccess && (
              <div className="mt-1 flex items-center gap-2">
                <p className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{ingestLag ?? "—"}</p>
                {ingestLag !== undefined && (
                  <Badge tone={ingestLag > 50 ? "warning" : "success"}>{ingestLag > 50 ? "Elevated" : "Nominal"}</Badge>
                )}
              </div>
            )}
          </CardBody>
        </Card>
      </div>

      <h2 className="mt-8 text-lg font-semibold text-slate-900 dark:text-slate-100">Stale Pending Transactions</h2>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Transactions still &ldquo;Pending&rdquo; more than 5 minutes after first observation — the worker has not yet
        seen a terminal RPC status for these.
      </p>
      <Card className="mt-3">
        <CardHeader>
          <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
            {staleTransactions.length} stale transaction{staleTransactions.length === 1 ? "" : "s"}
          </h3>
        </CardHeader>
        {pendingQuery.isPending && (
          <div className="p-5">
            <LoadingState rows={3} label="Loading pending transactions…" />
          </div>
        )}
        {pendingQuery.isError && (
          <div className="p-5">
            <ErrorState message="We could not load pending transactions." detail={pendingQuery.error.message} onRetry={() => pendingQuery.refetch()} />
          </div>
        )}
        {pendingQuery.isSuccess && staleTransactions.length === 0 && (
          <div className="p-5">
            <EmptyState message="No stale pending transactions right now." />
          </div>
        )}
        {pendingQuery.isSuccess && staleTransactions.length > 0 && (
          <DataTable columns={staleColumns(now)} rows={staleTransactions} getRowKey={(t) => t.transactionHash} />
        )}
      </Card>
    </div>
  );
}
