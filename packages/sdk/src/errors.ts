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
