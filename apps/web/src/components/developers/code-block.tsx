export function CodeBlock({ code, language = "ts" }: { code: string; language?: string }) {
  return (
    <pre className="overflow-x-auto rounded-md bg-slate-900 p-4 text-xs text-slate-100 dark:bg-black">
      <code data-language={language}>{code}</code>
    </pre>
  );
}
