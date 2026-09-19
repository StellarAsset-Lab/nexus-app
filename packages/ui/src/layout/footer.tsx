export interface FooterLink {
  readonly href: string;
  readonly label: string;
}

export interface FooterProps {
  readonly appName: string;
  readonly links: readonly FooterLink[];
}

export function Footer({ appName, links }: FooterProps) {
  return (
    <footer className="border-t border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950">
      <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-8 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-slate-500 dark:text-slate-400">
          &copy; {new Date().getFullYear()} {appName}
        </p>
        <nav className="flex flex-wrap gap-4" aria-label="Footer">
          {links.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100"
            >
              {link.label}
            </a>
          ))}
        </nav>
      </div>
    </footer>
  );
}
