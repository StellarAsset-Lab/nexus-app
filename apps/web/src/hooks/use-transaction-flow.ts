"use client";

import { useCallback, useState } from "react";
import type { AssembledTransaction, Result } from "@stellar/stellar-sdk/contract";
import { ContractError, executeTransaction, type TransactionLifecycleState } from "@nexus/sdk";
import { useWallet } from "@/components/wallet/wallet-context";

export interface TransactionFlowState {
  readonly state: TransactionLifecycleState;
  readonly error?: string;
  readonly hash?: string;
}

type Unwrapped<T> = T extends Result<infer U> ? U : T;

function isResult(value: unknown): value is Result<unknown> {
  return (
    typeof value === "object" &&
    value !== null &&
    "isOk" in value &&
    typeof (value as Result<unknown>).isOk === "function"
  );
}

/** Unwraps a Soroban `Result<T>` into `T`, throwing {@link ContractError} for `Err`. Values that aren't a `Result` pass through unchanged. */
function unwrapResult<T>(value: T): Unwrapped<T> {
  if (isResult(value)) {
    if (value.isErr()) {
      throw new ContractError(value.unwrapErr().message);
    }
    return value.unwrap() as Unwrapped<T>;
  }
  return value as Unwrapped<T>;
}

/**
 * Drives one write-transaction through the Sandbox: build (+ simulate) via
 * the caller's `buildTx`, then sign and submit through the connected
 * wallet, tracking the full {@link TransactionLifecycleState}. Every
 * Sandbox action (create/fund/settle/cancel/expire) shares this hook so
 * they all report state identically.
 */
export function useTransactionFlow<T>() {
  const wallet = useWallet();
  const [flow, setFlow] = useState<TransactionFlowState>({ state: "idle" });

  const run = useCallback(
    async (buildTx: () => Promise<AssembledTransaction<T>>): Promise<Unwrapped<T> | undefined> => {
      const signer = wallet.getSigner();
      if (!signer) {
        setFlow({
          state: "idle",
          error: wallet.networkMismatch
            ? "The connected wallet is on a different network than this app."
            : "Connect a wallet first.",
        });
        return undefined;
      }

      setFlow({ state: "building" });
      try {
        const tx = await buildTx();
        const sent = await executeTransaction(tx, signer, {
          onStateChange: (state) => setFlow((prev) => ({ ...prev, state })),
        });
        const result = unwrapResult(sent.result);
        setFlow({ state: "confirmed", hash: sent.hash });
        return result;
      } catch (cause) {
        setFlow({
          state: "failed",
          error: cause instanceof Error ? cause.message : "The transaction failed.",
        });
        return undefined;
      }
    },
    [wallet],
  );

  const reset = useCallback(() => setFlow({ state: "idle" }), []);

  return { flow, run, reset };
}
