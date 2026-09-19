import { AssembledTransaction, SentTransaction } from "@stellar/stellar-sdk/contract";
import type { Signer } from "@stellar/stellar-sdk/contract";
import { ConfirmationTimeoutError, SimulationError, SubmissionError, WalletRejectedError, RpcError } from "./errors";

/**
 * The full write-transaction lifecycle a caller may expose in its UI (see
 * spec §16). {@link executeTransaction} only drives the states from
 * "simulating" onward; "validating" and "building" happen in caller code
 * before an {@link AssembledTransaction} exists (local input validation, then
 * the generated contract client call that builds + initially simulates it).
 */
export type TransactionLifecycleState =
  | "idle"
  | "validating"
  | "building"
  | "simulating"
  | "awaiting-wallet"
  | "signed"
  | "submitting"
  | "pending"
  | "confirmed"
  | "failed"
  | "rejected"
  | "expired";

export interface TransactionLifecycleCallbacks {
  readonly onStateChange?: (state: TransactionLifecycleState) => void;
}

export interface TransactionResult<T> {
  readonly hash: string;
  readonly result: T;
}

/**
 * Re-simulates an already-built {@link AssembledTransaction} (to catch state
 * changes since it was first built), hands it to the connected wallet for
 * signing, submits it, and polls for confirmation — translating every SDK
 * failure mode into this SDK's typed error taxonomy.
 *
 * Confirmation submission and polling are delegated to the Stellar SDK's own
 * `AssembledTransaction.signAndSend`, which already retries `getTransaction`
 * with backoff for `MethodOptions.timeoutInSeconds` per spec §9.2 — this
 * function does not re-implement that loop, only translates its outcomes
 * using the SDK's own public `AssembledTransaction.Errors` /
 * `SentTransaction.Errors` class references.
 */
export async function executeTransaction<T>(
  tx: AssembledTransaction<T>,
  signer: Signer,
  callbacks: TransactionLifecycleCallbacks = {},
): Promise<TransactionResult<T>> {
  const emit = (state: TransactionLifecycleState) => callbacks.onStateChange?.(state);

  emit("simulating");
  try {
    await tx.simulate();
  } catch (cause) {
    emit("failed");
    throw new SimulationError("Transaction simulation failed.", { cause });
  }

  emit("awaiting-wallet");
  try {
    const sent = await tx.signAndSend({
      signTransaction: signer.signTransaction,
      watcher: {
        onSubmitted: () => emit("pending"),
        onProgress: () => emit("pending"),
      },
    });

    const hash = sent.sendTransactionResponse?.hash;
    if (hash === undefined) {
      emit("failed");
      throw new SubmissionError("The transaction was sent but Stellar RPC returned no transaction hash.");
    }

    emit("confirmed");
    return { hash, result: sent.result };
  } catch (cause) {
    if (cause instanceof SubmissionError) {
      throw cause;
    }

    if (cause instanceof AssembledTransaction.Errors.UserRejected) {
      emit("rejected");
      throw new WalletRejectedError("The wallet rejected the signing request.", { cause });
    }
    if (
      cause instanceof AssembledTransaction.Errors.SimulationFailed ||
      cause instanceof AssembledTransaction.Errors.ExpiredState ||
      cause instanceof AssembledTransaction.Errors.RestorationFailure ||
      cause instanceof AssembledTransaction.Errors.FakeAccount
    ) {
      emit("failed");
      throw new SimulationError("Transaction simulation failed during signing or submission.", { cause });
    }
    if (
      cause instanceof SentTransaction.Errors.SendFailed ||
      cause instanceof SentTransaction.Errors.SendResultOnly
    ) {
      emit("failed");
      throw new SubmissionError("Stellar RPC rejected the submitted transaction.", { cause });
    }
    if (cause instanceof SentTransaction.Errors.TransactionStillPending) {
      emit("expired");
      throw new ConfirmationTimeoutError(
        "The transaction did not reach a final status within the configured timeout. It may still confirm later; check its hash on the network directly.",
        { cause },
      );
    }

    emit("failed");
    throw new RpcError("Transaction submission or confirmation failed.", { cause });
  }
}
