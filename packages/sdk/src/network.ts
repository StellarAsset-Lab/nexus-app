import { Networks } from "@stellar/stellar-sdk";
import { ConfigurationError } from "./errors";
import type { NetworkConfig, NetworkName } from "./types";

const DEFAULT_TESTNET_RPC_URL = "https://soroban-testnet.stellar.org";
const DEFAULT_TESTNET_HORIZON_URL = "https://horizon-testnet.stellar.org";
const DEFAULT_MAINNET_HORIZON_URL = "https://horizon.stellar.org";

export interface ResolveNetworkConfigOptions {
  /**
   * Required for "mainnet": there is no default public Mainnet Soroban RPC
   * URL, so a configured third-party provider URL must be supplied. Optional
   * for "testnet", where it overrides the well-known default.
   */
  readonly rpcUrl?: string;
  readonly horizonUrl?: string;
}

export function isNetworkName(value: string): value is NetworkName {
  return value === "testnet" || value === "mainnet";
}

export function resolveNetworkConfig(
  network: NetworkName,
  options: ResolveNetworkConfigOptions = {},
): NetworkConfig {
  switch (network) {
    case "testnet":
      return {
        name: "testnet",
        networkPassphrase: Networks.TESTNET,
        rpcUrl: options.rpcUrl ?? DEFAULT_TESTNET_RPC_URL,
        horizonUrl: options.horizonUrl ?? DEFAULT_TESTNET_HORIZON_URL,
      };
    case "mainnet": {
      if (!options.rpcUrl) {
        throw new ConfigurationError(
          "Mainnet requires an explicit rpcUrl from a configured third-party provider; there is no default public Mainnet Soroban RPC URL.",
        );
      }
      return {
        name: "mainnet",
        networkPassphrase: Networks.PUBLIC,
        rpcUrl: options.rpcUrl,
        horizonUrl: options.horizonUrl ?? DEFAULT_MAINNET_HORIZON_URL,
      };
    }
  }
}
