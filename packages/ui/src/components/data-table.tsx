import type { ReactNode } from "react";

export interface DataTableColumn<T> {
  readonly key: string;
  readonly header: string;
  readonly render: (row: T) => ReactNode;
  readonly align?: "left" | "right" | "center";
}

export interface DataTableProps<T> {
  readonly columns: readonly DataTableColumn<T>[];
  readonly rows: readonly T[];
  readonly getRowKey: (row: T) => string;
  readonly onRowClick?: (row: T) => void;
}

const ALIGN_CLASSES = { left: "text-left", right: "text-right", center: "text-center" } as const;

/** A plain presentational table. Callers own data fetching, pagination, and filtering — this component only renders what it's given. */
export function DataTable<T>({ columns, rows, getRowKey, onRowClick }: DataTableProps<T>) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="border-b border-slate-200 dark:border-slate-800">
            {columns.map((column) => (
              <th
                key={column.key}
                className={[
                  "px-3 py-2 font-medium text-slate-500 dark:text-slate-400",
                  ALIGN_CLASSES[column.align ?? "left"],
                ].join(" ")}
                scope="col"
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={getRowKey(row)}
              onClick={onRowClick ? () => onRowClick(row) : undefined}
              className={[
                "border-b border-slate-100 dark:border-slate-800/60",
                onRowClick ? "cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800/50" : "",
              ].join(" ")}
            >
              {columns.map((column) => (
                <td key={column.key} className={["px-3 py-2", ALIGN_CLASSES[column.align ?? "left"]].join(" ")}>
                  {column.render(row)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
