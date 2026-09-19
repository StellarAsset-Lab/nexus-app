"use client";

import { useEffect, useRef, type ReactNode } from "react";

export interface ModalProps {
  readonly open: boolean;
  readonly onClose: () => void;
  readonly title: string;
  readonly children: ReactNode;
}

/** Uses the native <dialog> element for built-in focus trapping, Escape-to-close, and backdrop semantics. */
export function Modal({ open, onClose, title, children }: ModalProps) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    if (open && !dialog.open) {
      dialog.showModal();
    } else if (!open && dialog.open) {
      dialog.close();
    }
  }, [open]);

  return (
    <dialog
      ref={ref}
      onClose={onClose}
      onCancel={onClose}
      aria-labelledby="nexus-modal-title"
      className="rounded-lg border border-slate-200 bg-white p-0 backdrop:bg-slate-900/50 dark:border-slate-800 dark:bg-slate-900"
    >
      <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4 dark:border-slate-800">
        <h2 id="nexus-modal-title" className="text-sm font-semibold text-slate-900 dark:text-slate-100">
          {title}
        </h2>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="rounded p-1 text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
        >
          ✕
        </button>
      </div>
      <div className="px-5 py-4">{children}</div>
    </dialog>
  );
}
