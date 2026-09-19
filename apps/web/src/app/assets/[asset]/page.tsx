"use client";

import { use } from "react";
import Link from "next/link";
import { Badge, Card, CardBody, CardHeader, DataTable, EmptyState, ErrorState, LoadingState, StatusPill, type DataTableColumn } from "@nexus/ui";
import { useAsset, useAssetActivity } from "@/hooks/use-assets";
import type { OrderSummary } from "@/lib/api-types";
import { orderStatusTone } from "@/lib/status-tone";

const ORDER_COLUMNS: readonly DataTableColumn<OrderSummary>[] = [
  { key: "orderId", header: "Order", render: (o) => `#${o.orderId}` },
  { key: "status", header: "Status", render: (o) => <StatusPill tone={orderStatusTone(o.status)} label={o.status} /> },
  { key: "buyer", header: "Buyer", render: (o) => <span className="font-mono text-xs">{o.buyer}</span> },
  { key: "distributor", header: "Distributor", render: (o) => <span className="font-mono text-xs">{o.distributor}</span> },
  { key: "assetAmount", header: "Asset Amount", align: "right", render: (o) => o.assetAmount },
  { key: "paymentAmount", header: "Payment Amount", align: "right", render: (o) => o.paymentAmount },
  {
    key: "lastTransactionHash",
    header: "Last Transaction",
    render: (o) => (
      <Link href={`/transactions/${o.lastTransactionHash}`} className="font-mono text-xs text-blue-600 hover:underline dark:text-blue-400">
        {o.lastTransactionHash.slice(0, 12)}…
      </Link>
    ),
  },
];

export default function AssetDetailPage({ params }: { params: Promise<{ asset: string }> }) {
  const { asset } = use(params);
  const assetQuery = useAsset(asset);
  const activityQuery = useAssetActivity(asset);

  return (
    <div className="mx-auto max-w-7xl px-4 py-10">
      <Link href="/assets" className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100">
        ← All assets
      </Link>

      {assetQuery.isPending && (
        <div className="mt-6">
          <LoadingState rows={2} label="Loading asset…" />
        </div>
      )}
      {assetQuery.isError && (
        <div className="mt-6">
          <ErrorState
            message="We could not load this asset."
            detail={assetQuery.error.message}
            onRetry={() => assetQuery.refetch()}
          />
        </div>
      )}
      {assetQuery.isSuccess && (
        <Card className="mt-6">
          <CardHeader className="flex items-center justify-between">
            <div>
              <h1 className="font-mono text-lg font-semibold text-slate-900 dark:text-slate-100">{assetQuery.data.asset}</h1>
              <p className="mt-1 font-mono text-xs text-slate-500 dark:text-slate-400">Issuer: {assetQuery.data.issuer}</p>
            </div>
            <Badge tone={assetQuery.data.active ? "success" : "neutral"}>{assetQuery.data.active ? "Active" : "Inactive"}</Badge>
          </CardHeader>
          <CardBody className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-slate-500 dark:text-slate-400">First seen ledger</p>
              <p className="font-medium text-slate-900 dark:text-slate-100">{assetQuery.data.firstSeenLedger}</p>
            </div>
            <div>
              <p className="text-slate-500 dark:text-slate-400">Last updated ledger</p>
              <p className="font-medium text-slate-900 dark:text-slate-100">{assetQuery.data.lastUpdatedLedger}</p>
            </div>
          </CardBody>
        </Card>
      )}

      <h2 className="mt-8 text-lg font-semibold text-slate-900 dark:text-slate-100">Recent Orders</h2>
      <Card className="mt-3">
        {activityQuery.isPending && (
          <div className="p-5">
            <LoadingState rows={4} label="Loading order activity…" />
          </div>
        )}
        {activityQuery.isError && (
          <div className="p-5">
            <ErrorState
              message="We could not load order activity for this asset."
              detail={activityQuery.error.message}
              onRetry={() => activityQuery.refetch()}
            />
          </div>
        )}
        {activityQuery.isSuccess && activityQuery.data.orders.length === 0 && (
          <div className="p-5">
            <EmptyState message="No orders have referenced this asset yet." />
          </div>
        )}
        {activityQuery.isSuccess && activityQuery.data.orders.length > 0 && (
          <DataTable columns={ORDER_COLUMNS} rows={activityQuery.data.orders} getRowKey={(o) => String(o.orderId)} />
        )}
      </Card>
    </div>
  );
}
