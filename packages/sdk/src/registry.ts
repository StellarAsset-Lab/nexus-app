import type { AssetRecord, DistributionConfig, EligibilityRecord } from "@nexus/contracts-registry";
import type { NexusClient } from "./client";
import { ConfigurationError, RpcError } from "./errors";
import { assertValidStellarAddress } from "./validation";

function requireRegistryContract(client: NexusClient) {
  if (!client.registryContract) {
    throw new ConfigurationError(
      "The Registry contract is not configured (registryContractId is unset). Set NEXT_PUBLIC_REGISTRY_CONTRACT_ID once the Registry contract is deployed.",
    );
  }
  return client.registryContract;
}

/** Runs a generated read-only contract call and unwraps its simulated result, translating any failure into a typed {@link RpcError}. */
async function simulateRead<T>(call: () => Promise<{ result: T }>): Promise<T> {
  try {
    const { result } = await call();
    return result;
  } catch (cause) {
    throw new RpcError("Registry contract read failed.", { cause });
  }
}

/** Fetches the Registry record for `asset`, or `undefined` if the asset has never been registered. */
export async function getAsset(client: NexusClient, asset: string): Promise<AssetRecord | undefined> {
  const contract = requireRegistryContract(client);
  assertValidStellarAddress(asset, "asset");
  return simulateRead(() => contract.get_asset({ asset }));
}

/** Fetches the distribution config for `asset`/`distributor`, or `undefined` if none is registered. */
export async function getDistribution(
  client: NexusClient,
  asset: string,
  distributor: string,
): Promise<DistributionConfig | undefined> {
  const contract = requireRegistryContract(client);
  assertValidStellarAddress(asset, "asset");
  assertValidStellarAddress(distributor, "distributor");
  return simulateRead(() => contract.get_distribution({ asset, distributor }));
}

/** Fetches the eligibility record for `asset`/`distributor`/`buyer`, or `undefined` if none is set. */
export async function getEligibility(
  client: NexusClient,
  asset: string,
  distributor: string,
  buyer: string,
): Promise<EligibilityRecord | undefined> {
  const contract = requireRegistryContract(client);
  assertValidStellarAddress(asset, "asset");
  assertValidStellarAddress(distributor, "distributor");
  assertValidStellarAddress(buyer, "buyer");
  return simulateRead(() => contract.get_eligibility({ asset, distributor, buyer }));
}

/** Current on-chain eligibility check for `buyer` under `asset`/`distributor` (as of the latest ledger, not a cached projection). */
export async function isEligible(
  client: NexusClient,
  asset: string,
  distributor: string,
  buyer: string,
): Promise<boolean> {
  const contract = requireRegistryContract(client);
  assertValidStellarAddress(asset, "asset");
  assertValidStellarAddress(distributor, "distributor");
  assertValidStellarAddress(buyer, "buyer");
  return simulateRead(() => contract.is_eligible({ asset, distributor, buyer }));
}

export async function isAssetActive(client: NexusClient, asset: string): Promise<boolean> {
  const contract = requireRegistryContract(client);
  assertValidStellarAddress(asset, "asset");
  return simulateRead(() => contract.is_asset_active({ asset }));
}

export async function isDistributionActive(
  client: NexusClient,
  asset: string,
  distributor: string,
): Promise<boolean> {
  const contract = requireRegistryContract(client);
  assertValidStellarAddress(asset, "asset");
  assertValidStellarAddress(distributor, "distributor");
  return simulateRead(() => contract.is_distribution_active({ asset, distributor }));
}

export async function getRegistryAdmin(client: NexusClient): Promise<string> {
  const contract = requireRegistryContract(client);
  return simulateRead(() => contract.admin());
}

export async function getRegistryPendingAdmin(client: NexusClient): Promise<string | undefined> {
  const contract = requireRegistryContract(client);
  return simulateRead(() => contract.pending_admin());
}
