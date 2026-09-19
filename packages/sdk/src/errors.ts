/**
 * Base class for all typed Nexus SDK errors. Every error category the SDK
 * throws extends this so callers can distinguish Nexus-originated failures
 * from arbitrary exceptions with `instanceof NexusError`, and the original
 * cause (e.g. a network or RPC error) is preserved via the standard
 * `Error.cause` chain rather than swallowed.
 */
export abstract class NexusError extends Error {
  abstract readonly code: string;

  constructor(message: string, options?: { cause?: unknown }) {
    super(message, options);
    this.name = new.target.name;
  }
}

/** The SDK was configured with missing, invalid, or mutually inconsistent settings. */
export class ConfigurationError extends NexusError {
  readonly code = "CONFIGURATION_ERROR";
}

/** Caller-supplied input (an address, amount, ledger number, etc.) failed local validation before any network call was made. */
export class ValidationError extends NexusError {
  readonly code = "VALIDATION_ERROR";
}

/** A Stellar RPC or Horizon request failed (network error, timeout, non-2xx response). */
export class RpcError extends NexusError {
  readonly code = "RPC_ERROR";
}

/** Soroban simulation of a transaction failed (e.g. would-fail contract call, insufficient resources, expired state). */
export class SimulationError extends NexusError {
  readonly code = "SIMULATION_ERROR";
}

/** The connected wallet declined to sign, or the user cancelled the signing prompt. Distinct from a network failure. */
export class WalletRejectedError extends NexusError {
  readonly code = "WALLET_REJECTED_ERROR";
}

/** A signed transaction was rejected by Stellar RPC when submitted (e.g. bad sequence number, insufficient fee). */
export class SubmissionError extends NexusError {
  readonly code = "SUBMISSION_ERROR";
}

/** The transaction was submitted successfully but did not reach a final status within the configured timeout. It may still confirm later; this is not a failure, it is an unknown outcome. */
export class ConfirmationTimeoutError extends NexusError {
  readonly code = "CONFIRMATION_TIMEOUT_ERROR";
}

/** The contract itself returned a `Result::Err` (a `#[contracterror]` variant), as opposed to a network or simulation failure. */
export class ContractError extends NexusError {
  readonly code = "CONTRACT_ERROR";
}

/** The indexed (PostgreSQL) view of protocol state is known to be behind the chain. Distinct from a network/RPC failure. */
export class IndexerLagError extends NexusError {
  readonly code = "INDEXER_LAG_ERROR";
}
