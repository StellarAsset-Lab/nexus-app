import type { ReactNode } from "react";

export interface EmptyStateProps {
  readonly message: string;
  readonly action?: ReactNode;
}

/** For genuinely empty results — e.g. "No indexed Nexus activity yet." Never used to mask a fetch that hasn't run yet. */
export function EmptyState({ message, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed border-slate-300 px-6 py-10 text-center dark:border-slate-700">
      <p className="text-sm text-slate-500 dark:text-slate-400">{message}</p>
      {action}
    </div>
  );
}
