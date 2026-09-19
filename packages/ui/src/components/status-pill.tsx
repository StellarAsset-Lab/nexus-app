import type { HTMLAttributes } from "react";

/**
 * Generic status tones. Callers map their own domain status (order lifecycle,
 * component health, etc.) to one of these — this component has no knowledge
 * of order/protocol semantics, per the package's "no page-specific data
 * logic" rule.
 */
export type StatusTone = "positive" | "pending" | "negative" | "neutral" | "unknown";

export interface StatusPillProps extends HTMLAttributes<HTMLSpanElement> {
  readonly tone: StatusTone;
  readonly label: string;
}

const TONE_CLASSES: Record<StatusTone, string> = {
  positive: "bg-emerald-50 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300",
  pending: "bg-amber-50 text-amber-700 dark:bg-amber-950 dark:text-amber-300",
  negative: "bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300",
  neutral: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300",
  unknown: "bg-slate-50 text-slate-500 dark:bg-slate-900 dark:text-slate-400",
};

// Distinct marker shapes per tone so status is never conveyed by color alone.
const TONE_MARKERS: Record<StatusTone, string> = {
  positive: "●",
  pending: "◐",
  negative: "✕",
  neutral: "○",
  unknown: "?",
};

export function StatusPill({ tone, label, className, ...rest }: StatusPillProps) {
  return (
    <span
      className={[
        "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium",
        TONE_CLASSES[tone],
        className ?? "",
      ].join(" ")}
      {...rest}
    >
      <span aria-hidden="true">{TONE_MARKERS[tone]}</span>
      {label}
    </span>
  );
}
