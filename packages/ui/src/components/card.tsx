import type { HTMLAttributes, ReactNode } from "react";

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  readonly children: ReactNode;
}

export function Card({ children, className, ...rest }: CardProps) {
  return (
    <div
      className={[
        "rounded-lg border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900",
        className ?? "",
      ].join(" ")}
      {...rest}
    >
      {children}
    </div>
  );
}

export function CardHeader({ children, className, ...rest }: CardProps) {
  return (
    <div className={["border-b border-slate-200 px-5 py-4 dark:border-slate-800", className ?? ""].join(" ")} {...rest}>
      {children}
    </div>
  );
}

export function CardBody({ children, className, ...rest }: CardProps) {
  return (
    <div className={["px-5 py-4", className ?? ""].join(" ")} {...rest}>
      {children}
    </div>
  );
}
