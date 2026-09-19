"use client";

import { useWallet } from "./wallet-context";

function truncateAddress(address: string): string {
  return `${address.slice(0, 4)}…${address.slice(-4)}`;
}

export function ConnectWallet() {
  const { status, address, error, networkMismatch, connect, disconnect } = useWallet();

  if (status === "connected" && address !== undefined) {
    return (
      <div className="flex items-center gap-2 text-sm">
        {networkMismatch && (
          <span className="rounded border border-amber-500 px-2 py-1 text-amber-600 dark:text-amber-400" role="alert">
            Wallet network mismatch
          </span>
        )}
        <span className="font-mono" title={address}>
          {truncateAddress(address)}
        </span>
        <button
          type="button"
          onClick={() => void disconnect()}
          className="rounded border border-current px-3 py-1 text-sm"
        >
          Disconnect
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-end gap-1">
      <button
        type="button"
        onClick={() => void connect()}
        disabled={status === "connecting"}
        className="rounded border border-current px-3 py-1 text-sm disabled:opacity-50"
      >
        {status === "connecting" ? "Connecting…" : "Connect wallet"}
      </button>
      {status === "error" && error !== undefined && (
        <span className="text-xs text-red-600 dark:text-red-400" role="alert">
          {error}
        </span>
      )}
    </div>
  );
}
