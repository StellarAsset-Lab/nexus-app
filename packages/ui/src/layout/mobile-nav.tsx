"use client";

import { useState } from "react";
import type { NavItem } from "./header";

export interface MobileNavProps {
  readonly navItems: readonly NavItem[];
}

export function MobileNav({ navItems }: MobileNavProps) {
  const [open, setOpen] = useState(false);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        aria-expanded={open}
        aria-label={open ? "Close menu" : "Open menu"}
        className="flex h-9 w-9 items-center justify-center rounded-md border border-slate-300 dark:border-slate-700"
      >
        <span aria-hidden="true">{open ? "✕" : "☰"}</span>
      </button>
      {open && (
        <nav
          aria-label="Mobile"
          className="absolute right-0 top-11 z-10 w-48 rounded-md border border-slate-200 bg-white py-2 shadow-lg dark:border-slate-800 dark:bg-slate-900"
        >
          {navItems.map((item) => (
            <a
              key={item.href}
              href={item.href}
              onClick={() => setOpen(false)}
              className="block px-4 py-2 text-sm text-slate-700 hover:bg-slate-50 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {item.label}
            </a>
          ))}
        </nav>
      )}
    </div>
  );
}
