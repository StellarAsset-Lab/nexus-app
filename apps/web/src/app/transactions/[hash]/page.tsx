"use client";

import { use } from "react";
import Link from "next/link";
import {
  Badge,
  Card,
  CardBody,
  CardHeader,
  EmptyState,
  ErrorState,
  LoadingState,
  StatusPill,
  TransactionTimeline,
  type TimelineStep,
} from "@nexus/ui";
import { useOrder, useTransaction, useTransactionEvents } from "@/hooks/use-transactions";
import { orderStatusTone, transactionStatusTone } from "@/lib/status-tone";

export default function TransactionDetailPage({ params }: { params: Promise<{ hash: string }> }) {
  const { hash } = use(params);
  const txQuery = useTransaction(hash);
  const eventsQuery = useTransactionEvents(hash);
  const orderQuery = useOrder(txQuery.data?.orderId);

  const steps: TimelineStep[] =
    eventsQuery.data?.events.map((e) => ({
      key: e.eventId,
      label: e.eventType,
      status: e.decodeError ? "failed" : "complete",
      detail: `Ledger ${e.ledger}`,
    })) ?? [];

  return (
    <div className="mx-auto max-w-7xl px-4 py-10">
      <Link href="/transactions" className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100">
        ← All transactions
      </Link>

      {txQuery.isPending && (
        <div className="mt-6">
          <LoadingState rows={2} label="Loading transaction…" />
        </div>
      )}
      {txQuery.isError && (
        <div className="mt-6">
          <ErrorState message="We could not load this transaction." detail={txQuery.error.message} onRetry={() => txQuery.refetch()} />
        </div>
      )}
      {txQuery.isSuccess && (
        <>
          <Card className="mt-6">
            <CardHeader className="flex items-center justify-between">
              <h1 className="font-mono text-sm font-semibold text-slate-900 dark:text-slate-100">{txQuery.data.transactionHash}</h1>
              <StatusPill tone={transactionStatusTone(txQuery.data.status)} label={txQuery.data.status} />
            </CardHeader>
            <CardBody className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-3">
              <div>
                <p className="text-slate-500 dark:text-slate-400">Ledger</p>
                <p className="font-medium text-slate-900 dark:text-slate-100">{txQuery.data.ledger ?? "Not yet confirmed"}</p>
              </div>
              <div>
                <p className="text-slate-500 dark:text-slate-400">Source contract</p>
                <p className="font-mono text-xs font-medium text-slate-900 dark:text-slate-100">{txQuery.data.sourceContract ?? "—"}</p>
              </div>
              <div>
                <p className="text-slate-500 dark:text-slate-400">First observed</p>
                <p className="font-medium text-slate-900 dark:text-slate-100">{new Date(txQuery.data.firstObservedAt).toLocaleString()}</p>
              </div>
            </CardBody>
          </Card>

          {txQuery.data.orderId !== undefined && (
            <Card className="mt-6">
              <CardHeader>
                <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Related Order #{txQuery.data.orderId}</h2>
              </CardHeader>
              <CardBody>
                {orderQuery.isPending && <LoadingState rows={1} label="Loading order…" />}
                {orderQuery.isError && <ErrorState message="We could not load the related order." detail={orderQuery.error.message} />}
                {orderQuery.isSuccess && (
                  <div className="flex flex-wrap items-center gap-4 text-sm">
                    <StatusPill tone={orderStatusTone(orderQuery.data.status)} label={orderQuery.data.status} />
                    <Link href={`/assets/${encodeURIComponent(orderQuery.data.asset)}`} className="font-mono text-xs text-blue-600 hover:underline dark:text-blue-400">
                      {orderQuery.data.asset}
                    </Link>
                    <Badge tone="neutral">{orderQuery.data.assetAmount} for {orderQuery.data.paymentAmount}</Badge>
                  </div>
                )}
              </CardBody>
            </Card>
          )}

          <h2 className="mt-8 text-lg font-semibold text-slate-900 dark:text-slate-100">Event Timeline</h2>
          <Card className="mt-3">
            <CardBody>
              {eventsQuery.isPending && <LoadingState rows={3} label="Loading events…" />}
              {eventsQuery.isError && (
                <ErrorState message="We could not load this transaction's events." detail={eventsQuery.error.message} onRetry={() => eventsQuery.refetch()} />
              )}
              {eventsQuery.isSuccess && steps.length === 0 && <EmptyState message="No indexed contract events for this transaction." />}
              {eventsQuery.isSuccess && steps.length > 0 && <TransactionTimeline steps={steps} />}
            </CardBody>
          </Card>
        </>
      )}
    </div>
  );
}
