import { Button } from "./button";

export interface ErrorStateProps {
  readonly message?: string;
  readonly onRetry?: () => void;
  /** The raw underlying error, shown only in a collapsible details panel for developers — never in the primary message. */
  readonly detail?: string;
}

export function ErrorState({ message = "We could not load this data.", onRetry, detail }: ErrorStateProps) {
  return (
    <div role="alert" className="flex flex-col items-center gap-3 rounded-lg border border-red-200 px-6 py-10 text-center dark:border-red-900">
      <p className="text-sm text-red-700 dark:text-red-400">{message}</p>
      {onRetry && (
        <Button variant="secondary" size="sm" onClick={onRetry}>
          Retry
        </Button>
      )}
      {detail && (
        <details className="mt-2 max-w-full text-left text-xs text-slate-500 dark:text-slate-400">
          <summary className="cursor-pointer">Details</summary>
          <pre className="mt-1 overflow-x-auto whitespace-pre-wrap">{detail}</pre>
        </details>
      )}
    </div>
  );
}
