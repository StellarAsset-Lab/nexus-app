import type { AssembledTransaction, MethodOptions, Result } from "@stellar/stellar-sdk/contract";
import type { OrderRecord } from "@nexus/contracts-order";
import type { NexusClient } from "./client";
import { ConfigurationError, RpcError, ValidationError } from "./errors";
import { assertValidStellarAddress } from "./validation";

function requireOrderContract(client: NexusClient) {
  if (!client.orderContract) {
    throw new ConfigurationError(
      "The Order contract is not configured (orderContractId is unset). Set NEXT_PUBLIC_ORDER_CONTRACT_ID once the Order contract is deployed.",
    );
  }
  return client.orderContract;
}

/** Runs a generated read-only contract call and unwraps its simulated result, translating any failure into a typed {@link RpcError}. */
async function simulateRead<T>(call: () => Promise<{ result: T }>): Promise<T> {
  try {
    const { result } = await call();
    return result;
  } catch (cause) {
    throw new RpcError("Order contract read failed.", { cause });
  }
}

function assertValidOrderId(orderId: bigint): bigint {
  if (orderId < 0n) {
    throw new ValidationError(`orderId must be a non-negative integer, got ${orderId}`);
  }
  return orderId;
}

/** Fetches the order record for `orderId`, or `undefined` if no such order exists. */
export async function getOrder(client: NexusClient, orderId: bigint): Promise<OrderRecord | undefined> {
  const contract = requireOrderContract(client);
  assertValidOrderId(orderId);
  return simulateRead(() => contract.get_order({ order_id: orderId }));
}

export async function orderExists(client: NexusClient, orderId: bigint): Promise<boolean> {
  const contract = requireOrderContract(client);
  assertValidOrderId(orderId);
  return simulateRead(() => contract.order_exists({ order_id: orderId }));
}

/** The order ID that will be assigned to the next created order. */
export async function nextOrderId(client: NexusClient): Promise<bigint> {
  const contract = requireOrderContract(client);
  return simulateRead(() => contract.next_order_id());
}

/** Whether the Order gateway currently blocks new order creation and funding (settlement, cancellation, and expiry remain available). */
export async function isPaused(client: NexusClient): Promise<boolean> {
  const contract = requireOrderContract(client);
  return simulateRead(() => contract.is_paused());
}

export async function getOrderAdmin(client: NexusClient): Promise<string> {
  const contract = requireOrderContract(client);
  return simulateRead(() => contract.admin());
}

export async function getOrderPendingAdmin(client: NexusClient): Promise<string | undefined> {
  const contract = requireOrderContract(client);
  return simulateRead(() => contract.pending_admin());
}

/** The Registry contract address this Order contract was initialized against. */
export async function getOrderRegistry(client: NexusClient): Promise<string> {
  const contract = requireOrderContract(client);
  return simulateRead(() => contract.registry());
}

function assertPositiveAmount(value: bigint, fieldName: string): bigint {
  if (value <= 0n) {
    throw new ValidationError(`${fieldName} must be a positive amount, got ${value}`);
  }
  return value;
}

export interface CreateOrderParams {
  readonly distributor: string;
  readonly buyer: string;
  readonly asset: string;
  readonly paymentAsset: string;
  /** The exact i128 token amount, as a bigint — never a float, per spec §7. */
  readonly assetAmount: bigint;
  readonly paymentAmount: bigint;
  readonly expiresAtLedger: number;
}

/**
 * Builds (and simulates) a `create_order` transaction. Returns the raw
 * {@link AssembledTransaction} — callers hand it to `executeTransaction`
 * with a connected wallet's {@link Signer} to actually sign and submit it.
 * This function never signs or submits anything itself.
 */
export function buildCreateOrder(
  client: NexusClient,
  params: CreateOrderParams,
  options?: MethodOptions,
): Promise<AssembledTransaction<Result<bigint>>> {
  const contract = requireOrderContract(client);
  assertValidStellarAddress(params.distributor, "distributor");
  assertValidStellarAddress(params.buyer, "buyer");
  assertValidStellarAddress(params.asset, "asset");
  assertValidStellarAddress(params.paymentAsset, "paymentAsset");
  assertPositiveAmount(params.assetAmount, "assetAmount");
  assertPositiveAmount(params.paymentAmount, "paymentAmount");
  if (params.expiresAtLedger <= 0) {
    throw new ValidationError(`expiresAtLedger must be a positive ledger sequence, got ${params.expiresAtLedger}`);
  }

  return contract.create_order(
    {
      distributor: params.distributor,
      buyer: params.buyer,
      asset: params.asset,
      payment_asset: params.paymentAsset,
      asset_amount: params.assetAmount,
      payment_amount: params.paymentAmount,
      expires_at_ledger: params.expiresAtLedger,
    },
    options,
  );
}

/** Builds (and simulates) a `fund_payment` transaction for `orderId`. */
export function buildFundPayment(client: NexusClient, orderId: bigint, options?: MethodOptions): Promise<AssembledTransaction<Result<void>>> {
  const contract = requireOrderContract(client);
  assertValidOrderId(orderId);
  return contract.fund_payment({ order_id: orderId }, options);
}

/** Builds (and simulates) a `fund_asset` transaction for `orderId`. */
export function buildFundAsset(client: NexusClient, orderId: bigint, options?: MethodOptions): Promise<AssembledTransaction<Result<void>>> {
  const contract = requireOrderContract(client);
  assertValidOrderId(orderId);
  return contract.fund_asset({ order_id: orderId }, options);
}

/** Builds (and simulates) a `settle` transaction for `orderId`. */
export function buildSettle(client: NexusClient, orderId: bigint, options?: MethodOptions): Promise<AssembledTransaction<Result<void>>> {
  const contract = requireOrderContract(client);
  assertValidOrderId(orderId);
  return contract.settle({ order_id: orderId }, options);
}

/** Builds (and simulates) a `cancel_order` transaction for `orderId`. */
export function buildCancelOrder(client: NexusClient, orderId: bigint, options?: MethodOptions): Promise<AssembledTransaction<Result<void>>> {
  const contract = requireOrderContract(client);
  assertValidOrderId(orderId);
  return contract.cancel_order({ order_id: orderId }, options);
}

/** Builds (and simulates) an `expire_order` transaction for `orderId`. */
export function buildExpireOrder(client: NexusClient, orderId: bigint, options?: MethodOptions): Promise<AssembledTransaction<Result<void>>> {
  const contract = requireOrderContract(client);
  assertValidOrderId(orderId);
  return contract.expire_order({ order_id: orderId }, options);
}
