"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { KitEventType, Networks as KitNetworks, StellarWalletsKit } from "@creit.tech/stellar-wallets-kit";
import type { Signer } from "@stellar/stellar-sdk/contract";
import { getPublicEnv } from "@/lib/env";
import { createWalletModules } from "@/lib/wallet-modules";
import { WalletContext, type WalletConnectionStatus, type WalletContextValue } from "@/components/wallet/wallet-context";

const WALLET_ID_STORAGE_KEY = "nexus.walletId";

function toKitNetwork(networkPassphrase: string): KitNetworks | undefined {
  return Object.values(KitNetworks).includes(networkPassphrase as KitNetworks)
    ? (networkPassphrase as KitNetworks)
    : undefined;
}

/** Reads a previously-selected wallet id from localStorage. Never throws: private browsing / blocked storage just means no persisted selection. */
function readPersistedWalletId(): string | undefined {
  try {
    return localStorage.getItem(WALLET_ID_STORAGE_KEY) ?? undefined;
  } catch {
    return undefined;
  }
}

function persistWalletId(id: string | undefined): void {
  try {
    if (id === undefined) {
      localStorage.removeItem(WALLET_ID_STORAGE_KEY);
    } else {
      localStorage.setItem(WALLET_ID_STORAGE_KEY, id);
    }
  } catch {
    // Best-effort only; the wallet still works within this session.
  }
}

export function WalletProvider({ children }: { children: React.ReactNode }) {
  const env = getPublicEnv();
  const [status, setStatus] = useState<WalletConnectionStatus>("disconnected");
  const [address, setAddress] = useState<string>();
  const [error, setError] = useState<string>();
  const [networkMismatch, setNetworkMismatch] = useState(false);
  const initialized = useRef(false);

  useEffect(() => {
    if (initialized.current) {
      return;
    }
    initialized.current = true;

    const kitNetwork = toKitNetwork(env.networkConfig.networkPassphrase);
    const persistedWalletId = readPersistedWalletId();

    StellarWalletsKit.init({
      modules: createWalletModules(),
      // The configured network passphrase isn't always one of the Wallets
      // Kit's known networks (e.g. a private Standalone/Sandbox network) —
      // the kit still works, it just can't pre-select one for wallets that
      // ask for it, so `network` is omitted entirely in that case.
      ...(kitNetwork !== undefined && { network: kitNetwork }),
      ...(persistedWalletId !== undefined && { selectedWalletId: persistedWalletId }),
    });

    const offStateUpdated = StellarWalletsKit.on(KitEventType.STATE_UPDATED, (event) => {
      const walletAddress = event.payload.address;
      if (walletAddress === undefined) {
        setStatus("disconnected");
        setAddress(undefined);
        setNetworkMismatch(false);
        return;
      }
      setAddress(walletAddress);
      setStatus("connected");
      setError(undefined);
      setNetworkMismatch(event.payload.networkPassphrase !== env.networkConfig.networkPassphrase);
    });

    const offDisconnect = StellarWalletsKit.on(KitEventType.DISCONNECT, () => {
      setStatus("disconnected");
      setAddress(undefined);
      setNetworkMismatch(false);
      persistWalletId(undefined);
    });

    return () => {
      offStateUpdated();
      offDisconnect();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- kit init must run exactly once per app lifetime
  }, []);

  const connect = useCallback(async () => {
    setStatus("connecting");
    setError(undefined);
    try {
      const { address: connectedAddress } = await StellarWalletsKit.authModal();
      setAddress(connectedAddress);
      setStatus("connected");
      persistWalletId(StellarWalletsKit.selectedModule?.productId);

      const { networkPassphrase } = await StellarWalletsKit.getNetwork();
      setNetworkMismatch(networkPassphrase !== env.networkConfig.networkPassphrase);
    } catch (cause) {
      setStatus("error");
      setError(cause instanceof Error ? cause.message : "Failed to connect wallet.");
    }
  }, [env.networkConfig.networkPassphrase]);

  const disconnect = useCallback(async () => {
    try {
      await StellarWalletsKit.disconnect();
    } finally {
      setStatus("disconnected");
      setAddress(undefined);
      setNetworkMismatch(false);
      persistWalletId(undefined);
    }
  }, []);

  const getSigner = useCallback((): Signer | undefined => {
    if (status !== "connected" || address === undefined || networkMismatch) {
      return undefined;
    }
    return {
      address,
      signTransaction: async (xdr, opts) => {
        try {
          const result = await StellarWalletsKit.signTransaction(xdr, {
            ...opts,
            networkPassphrase: env.networkConfig.networkPassphrase,
            address,
          });
          return result;
        } catch (cause) {
          // The Wallets Kit rejects its signTransaction promise on failure
          // rather than returning the SEP-43 `{ error }` shape our SDK's
          // AssembledTransaction.handleWalletError expects. By the time this
          // is called the wallet is already connected, so a thrown error
          // here is almost always the user declining the signing prompt —
          // code -4 maps to the SDK's UserRejectedError, which the app
          // surfaces distinctly from a network/submission failure.
          return {
            signedTxXdr: "",
            error: {
              code: -4,
              message: cause instanceof Error ? cause.message : "The wallet declined to sign the transaction.",
            },
          };
        }
      },
    };
  }, [status, address, networkMismatch, env.networkConfig.networkPassphrase]);

  const value = useMemo<WalletContextValue>(
    () => ({
      status,
      networkMismatch,
      connect,
      disconnect,
      getSigner,
      ...(address !== undefined && { address }),
      ...(error !== undefined && { error }),
    }),
    [status, address, error, networkMismatch, connect, disconnect, getSigner],
  );

  return <WalletContext.Provider value={value}>{children}</WalletContext.Provider>;
}
