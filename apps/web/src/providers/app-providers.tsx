"use client";

import { QueryProvider } from "./query-provider";
import { WalletProvider } from "./wallet-provider";

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      <WalletProvider>{children}</WalletProvider>
    </QueryProvider>
  );
}
