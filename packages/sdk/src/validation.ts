import { Address } from "@stellar/stellar-sdk";
import { ValidationError } from "./errors";

/**
 * Validates that `value` is a well-formed Stellar address (account "G...",
 * contract "C...", or other strkey-encoded address) using the SDK's own
 * parser rather than a hand-rolled regex. Throws {@link ValidationError} on
 * failure so callers get a clear local error instead of a confusing RPC
 * failure from submitting garbage input.
 */
export function assertValidStellarAddress(value: string, fieldName: string): string {
  try {
    Address.fromString(value);
  } catch (cause) {
    throw new ValidationError(`${fieldName} is not a valid Stellar address: "${value}"`, { cause });
  }
  return value;
}

export function isValidStellarAddress(value: string): boolean {
  try {
    Address.fromString(value);
    return true;
  } catch {
    return false;
  }
}
