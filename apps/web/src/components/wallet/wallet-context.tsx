"use client";

import { createContext, useContext } from "react";
import type { Signer } from "@stellar/stellar-sdk/contract";

export type WalletConnectionStatus = "disconnected" | "connecting" | "connected" | "error";

export interface WalletState {
  readonly status: WalletConnectionStatus;
  readonly address?: string;
  /** Set when status is "error", or when the wallet reports rejecting the last request. */
  readonly error?: string;
  /**
   * True once the connected wallet's own reported network passphrase is
   * known to differ from this application's configured network. The wallet
   * remains connected (so the mismatch can be surfaced clearly) but callers
   * must not sign or submit transactions while this is true — see spec §11.
   */
  readonly networkMismatch: boolean;
}

export interface WalletContextValue extends WalletState {
  connect(): Promise<void>;
  disconnect(): Promise<void>;
  /** A Signer usable directly with @nexus/sdk's `executeTransaction`. Undefined until connected without a network mismatch. */
  getSigner(): Signer | undefined;
}

export const WalletContext = createContext<WalletContextValue | undefined>(undefined);

export function useWallet(): WalletContextValue {
  const value = useContext(WalletContext);
  if (!value) {
    throw new Error("useWallet must be used within a <WalletProvider>.");
  }
  return value;
}
