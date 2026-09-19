"use client";

import { Badge, Card, CardBody, CardHeader, ErrorState, LoadingState, StatusPill } from "@nexus/ui";
import { useNetwork, useStatus } from "@/hooks/use-status";
import { componentHealthTone } from "@/lib/status-tone";

export default function StatusPage() {
  const statusQuery = useStatus();
  const networkQuery = useNetwork();

  const overall = statusQuery.data
    ? statusQuery.data.components.every((c) => c.status === "Operational")
      ? "All systems operational"
      : statusQuery.data.components.some((c) => c.status === "Unavailable")
        ? "Degraded — one or more systems unavailable"
        : "Degraded"
    : undefined;

  return (
    <div className="mx-auto max-w-4xl px-4 py-10">
      <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">System Status</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Live, directly-observed health of every Nexus dependency — checked on every page load, never a static claim.
      </p>

      {overall && (
        <div className="mt-4">
          <Badge tone={overall.startsWith("All") ? "success" : "warning"}>{overall}</Badge>
        </div>
      )}

      <Card className="mt-6">
        <CardHeader>
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Components</h2>
        </CardHeader>
        <CardBody className="flex flex-col gap-3">
          {statusQuery.isPending && <LoadingState rows={3} label="Checking component status…" />}
          {statusQuery.isError && (
            <ErrorState message="We could not load system status." detail={statusQuery.error.message} onRetry={() => statusQuery.refetch()} />
          )}
          {statusQuery.isSuccess &&
            statusQuery.data.components.map((c) => (
              <div key={c.name} className="flex items-center justify-between border-b border-slate-100 pb-3 last:border-0 last:pb-0 dark:border-slate-800">
                <div>
                  <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{c.name}</p>
                  {c.detail && <p className="text-xs text-slate-500 dark:text-slate-400">{c.detail}</p>}
                </div>
                <StatusPill tone={componentHealthTone(c.status)} label={c.status} />
              </div>
            ))}
        </CardBody>
      </Card>

      <Card className="mt-6">
        <CardHeader>
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Network</h2>
        </CardHeader>
        <CardBody className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
          {networkQuery.isPending && <LoadingState rows={1} label="Loading network info…" />}
          {networkQuery.isError && (
            <div className="col-span-full">
              <ErrorState message="We could not load network info." detail={networkQuery.error.message} onRetry={() => networkQuery.refetch()} />
            </div>
          )}
          {networkQuery.isSuccess && (
            <>
              <div>
                <p className="text-slate-500 dark:text-slate-400">Network</p>
                <p className="font-medium text-slate-900 dark:text-slate-100">{networkQuery.data.network}</p>
              </div>
              <div>
                <p className="text-slate-500 dark:text-slate-400">RPC</p>
                <StatusPill tone={networkQuery.data.rpcHealthy ? "positive" : "negative"} label={networkQuery.data.rpcHealthy ? "Healthy" : "Unreachable"} />
              </div>
              <div>
                <p className="text-slate-500 dark:text-slate-400">Latest ledger</p>
                <p className="font-medium text-slate-900 dark:text-slate-100">{networkQuery.data.latestLedger ?? "—"}</p>
              </div>
              <div>
                <p className="text-slate-500 dark:text-slate-400">Indexer checkpoint</p>
                <p className="font-medium text-slate-900 dark:text-slate-100">{networkQuery.data.indexerCheckpoint ?? "None yet"}</p>
              </div>
            </>
          )}
        </CardBody>
      </Card>
    </div>
  );
}
