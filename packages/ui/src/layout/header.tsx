import type { ReactNode } from "react";

export interface NavItem {
  readonly href: string;
  readonly label: string;
}

export interface HeaderProps {
  readonly appName: string;
  readonly navItems: readonly NavItem[];
  readonly logo?: ReactNode;
  /** App-specific content (network indicator, connect-wallet button, etc.) — this package has no knowledge of wallets or network state. */
  readonly rightSlot?: ReactNode;
  readonly mobileNavSlot?: ReactNode;
}

export function Header({ appName, navItems, logo, rightSlot, mobileNavSlot }: HeaderProps) {
  return (
    <header className="border-b border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-4 px-4">
        <div className="flex items-center gap-6">
          <a href="/" className="flex items-center gap-2 font-semibold text-slate-900 dark:text-slate-100">
            {logo}
            {appName}
          </a>
          <nav className="hidden items-center gap-4 md:flex" aria-label="Primary">
            {navItems.map((item) => (
              <a
                key={item.href}
                href={item.href}
                className="text-sm text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100"
              >
                {item.label}
              </a>
            ))}
          </nav>
        </div>
        <div className="flex items-center gap-3">
          {rightSlot}
          <div className="md:hidden">{mobileNavSlot}</div>
        </div>
      </div>
    </header>
  );
}
