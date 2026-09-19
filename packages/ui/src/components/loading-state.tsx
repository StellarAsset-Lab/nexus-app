export interface LoadingStateProps {
  readonly rows?: number;
  readonly label?: string;
}

/** Reserves layout space with skeleton rows instead of a spinner, so content doesn't jump on load. */
export function LoadingState({ rows = 3, label = "Loading…" }: LoadingStateProps) {
  return (
    <div role="status" aria-label={label} className="flex flex-col gap-2">
      {Array.from({ length: rows }, (_, index) => (
        <div
          key={index}
          className="h-10 w-full animate-pulse rounded-md bg-slate-100 dark:bg-slate-800"
          aria-hidden="true"
        />
      ))}
      <span className="sr-only">{label}</span>
    </div>
  );
}
