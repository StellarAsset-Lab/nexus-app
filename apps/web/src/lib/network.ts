import type { NetworkConfig } from "@nexus/sdk";
import { getPublicEnv } from "./env";

/**
 * The application's single source of truth for "which network are we on".
 * Every place that needs the current network config (status displays,
 * explorer links, SDK client construction) should go through this module
 * rather than re-deriving it, so Testnet and Mainnet can never be mixed up.
 */
export function getNetworkConfig(): NetworkConfig {
  return getPublicEnv().networkConfig;
}

const EXPLORER_BASE_URL: Record<NetworkConfig["name"], string> = {
  testnet: "https://stellar.expert/explorer/testnet",
  mainnet: "https://stellar.expert/explorer/public",
};

/** Builds a block explorer URL from a real, confirmed transaction hash. Never call with a fabricated hash. */
export function buildTransactionExplorerUrl(transactionHash: string): string {
  const network = getNetworkConfig();
  return `${EXPLORER_BASE_URL[network.name]}/tx/${transactionHash}`;
}

/** Builds a block explorer URL for a Soroban contract address. */
export function buildContractExplorerUrl(contractId: string): string {
  const network = getNetworkConfig();
  return `${EXPLORER_BASE_URL[network.name]}/contract/${contractId}`;
}
