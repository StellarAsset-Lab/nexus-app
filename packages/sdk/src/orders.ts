import type { OrderRecord } from "@nexus/contracts-order";
import type { NexusClient } from "./client";
import { ConfigurationError, RpcError, ValidationError } from "./errors";

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
