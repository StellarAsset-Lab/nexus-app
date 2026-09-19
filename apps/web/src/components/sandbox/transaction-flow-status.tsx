import Link from "next/link";
import { Badge, StatusPill } from "@nexus/ui";
import type { TransactionFlowState } from "@/hooks/use-transaction-flow";
import { transactionStatusTone } from "@/lib/status-tone";

const STATE_LABEL: Record<TransactionFlowState["state"], string> = {
  idle: "Idle",
  validating: "Validating",
  building: "Building",
  simulating: "Simulating",
  "awaiting-wallet": "Awaiting wallet signature",
  signed: "Signed",
  submitting: "Submitting",
  pending: "Pending confirmation",
  confirmed: "Confirmed",
  failed: "Failed",
  rejected: "Rejected by wallet",
  expired: "Confirmation timed out",
};

/** Shared lifecycle display for every Sandbox action — never shows a synthetic "success" state before `executeTransaction` actually reports one. */
export function TransactionFlowStatus({ flow }: { flow: TransactionFlowState }) {
  if (flow.state === "idle" && flow.error === undefined) {
    return null;
  }

  return (
    <div className="mt-3 flex flex-col gap-2 text-sm">
      <StatusPill tone={transactionStatusTone(flow.state === "confirmed" ? "Confirmed" : flow.state === "failed" || flow.state === "rejected" ? "Failed" : "Pending")} label={STATE_LABEL[flow.state]} />
      {flow.error && (
        <p role="alert" className="text-red-600 dark:text-red-400">
          {flow.error}
        </p>
      )}
      {flow.hash && (
        <p>
          <Link href={`/transactions/${encodeURIComponent(flow.hash)}`} className="font-mono text-xs text-blue-600 hover:underline dark:text-blue-400">
            {flow.hash}
          </Link>
          <Badge tone="neutral" className="ml-2">
            View in Transaction Center
          </Badge>
        </p>
      )}
    </div>
  );
}
