export type TimelineStepStatus = "complete" | "current" | "pending" | "failed";

export interface TimelineStep {
  readonly key: string;
  readonly label: string;
  readonly status: TimelineStepStatus;
  /** e.g. a real transaction hash or ledger number. Never fabricate this — omit it while unknown. */
  readonly detail?: string;
}

export interface TransactionTimelineProps {
  readonly steps: readonly TimelineStep[];
}

const MARKER_CLASSES: Record<TimelineStepStatus, string> = {
  complete: "border-emerald-500 bg-emerald-500 text-white",
  current: "border-blue-500 bg-white text-blue-500 dark:bg-slate-900",
  pending: "border-slate-300 bg-white text-slate-400 dark:border-slate-700 dark:bg-slate-900",
  failed: "border-red-500 bg-red-500 text-white",
};

const MARKER_SYMBOL: Record<TimelineStepStatus, string> = {
  complete: "✓",
  current: "●",
  pending: "○",
  failed: "✕",
};

/** Renders a real transaction/order lifecycle as a vertical timeline. Callers supply the step labels and statuses — this component holds no protocol knowledge. */
export function TransactionTimeline({ steps }: TransactionTimelineProps) {
  return (
    <ol className="flex flex-col">
      {steps.map((step, index) => (
        <li key={step.key} className="flex gap-3">
          <div className="flex flex-col items-center">
            <span
              className={["flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 text-xs", MARKER_CLASSES[step.status]].join(" ")}
              aria-hidden="true"
            >
              {MARKER_SYMBOL[step.status]}
            </span>
            {index < steps.length - 1 && (
              <span className="my-1 w-px grow bg-slate-200 dark:bg-slate-800" aria-hidden="true" />
            )}
          </div>
          <div className="pb-6">
            <p className="text-sm font-medium text-slate-900 dark:text-slate-100">
              {step.label}
              <span className="sr-only"> — {step.status}</span>
            </p>
            {step.detail && <p className="text-xs text-slate-500 dark:text-slate-400">{step.detail}</p>}
          </div>
        </li>
      ))}
    </ol>
  );
}
