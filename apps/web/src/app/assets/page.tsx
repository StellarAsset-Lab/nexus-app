"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Badge, Card, DataTable, EmptyState, ErrorState, LoadingState, type DataTableColumn } from "@nexus/ui";
import { useAssets } from "@/hooks/use-assets";
import type { AssetResponse } from "@/lib/api-types";

const PAGE_LIMIT = 25;

const COLUMNS: readonly DataTableColumn<AssetResponse>[] = [
  { key: "asset", header: "Asset", render: (a) => <span className="font-mono text-xs">{a.asset}</span> },
  { key: "issuer", header: "Issuer", render: (a) => <span className="font-mono text-xs">{a.issuer}</span> },
  {
    key: "active",
    header: "Status",
    render: (a) => <Badge tone={a.active ? "success" : "neutral"}>{a.active ? "Active" : "Inactive"}</Badge>,
  },
  { key: "firstSeenLedger", header: "First Seen", align: "right", render: (a) => a.firstSeenLedger },
  { key: "lastUpdatedLedger", header: "Last Updated", align: "right", render: (a) => a.lastUpdatedLedger },
];

export default function AssetsPage() {
  const router = useRouter();
  const [activeFilter, setActiveFilter] = useState<boolean | undefined>(undefined);
  const [cursors, setCursors] = useState<readonly string[]>([]);
  const cursor = cursors[cursors.length - 1];

  const query = useAssets({
    ...(activeFilter !== undefined && { active: activeFilter }),
    ...(cursor !== undefined && { cursor }),
    limit: PAGE_LIMIT,
  });

  const filterOptions: readonly { label: string; value: boolean | undefined }[] = [
    { label: "All", value: undefined },
    { label: "Active", value: true },
    { label: "Inactive", value: false },
  ];

  return (
    <div className="mx-auto max-w-7xl px-4 py-10">
      <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">Asset Explorer</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Every Stellar asset the Registry contract has ever listed, as observed by the indexer — never a fabricated
        or estimated list.
      </p>

      <div className="mt-6 flex gap-2">
        {filterOptions.map((option) => (
          <button
            key={option.label}
            onClick={() => {
              setActiveFilter(option.value);
              setCursors([]);
            }}
            className={[
              "rounded-md px-3 py-1.5 text-sm font-medium",
              activeFilter === option.value
                ? "bg-slate-900 text-white dark:bg-slate-100 dark:text-slate-900"
                : "border border-slate-300 text-slate-700 hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800",
            ].join(" ")}
          >
            {option.label}
          </button>
        ))}
      </div>

      <Card className="mt-6">
        {query.isPending && (
          <div className="p-5">
            <LoadingState rows={5} label="Loading assets…" />
          </div>
        )}
        {query.isError && (
          <div className="p-5">
            <ErrorState
              message="We could not load the asset list."
              detail={query.error.message}
              onRetry={() => query.refetch()}
            />
          </div>
        )}
        {query.isSuccess && query.data.assets.length === 0 && (
          <div className="p-5">
            <EmptyState message="No indexed assets yet." />
          </div>
        )}
        {query.isSuccess && query.data.assets.length > 0 && (
          <>
            <DataTable
              columns={COLUMNS}
              rows={query.data.assets}
              getRowKey={(a) => a.asset}
              onRowClick={(a) => router.push(`/assets/${encodeURIComponent(a.asset)}`)}
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
