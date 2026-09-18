import { ConfigurationError, isNetworkName, resolveNetworkConfig } from "@nexus/sdk";
import type { NetworkConfig, NetworkName } from "@nexus/sdk";
import { z } from "zod";

/**
 * Centralized, validated access to browser-safe environment configuration.
 * Nothing else in the frontend should read `process.env.NEXT_PUBLIC_*`
 * directly (see docs/local-development.md and the project system prompt,
 * "do not access process.env directly throughout the frontend").
 *
 * There is no server-only counterpart here: per the application boundary,
 * the Next.js app never talks to Postgres or holds server secrets directly
 * (DATABASE_URL, API_KEY_PEPPER, etc. belong to the Go services only).
 */

const CONTRACT_ID_PLACEHOLDER = "<SET_AFTER_DEPLOYMENT>";

const rawEnvSchema = z.object({
  NEXT_PUBLIC_APP_NAME: z.string().min(1),
  NEXT_PUBLIC_NETWORK: z.string().refine(isNetworkName, {
    message: 'must be "testnet" or "mainnet"',
  }),
  NEXT_PUBLIC_NETWORK_PASSPHRASE: z.string().min(1),
  NEXT_PUBLIC_SOROBAN_RPC_URL: z.url(),
  NEXT_PUBLIC_HORIZON_URL: z.url(),
  NEXT_PUBLIC_REGISTRY_CONTRACT_ID: z.string().optional(),
  NEXT_PUBLIC_ORDER_CONTRACT_ID: z.string().optional(),
  NEXT_PUBLIC_API_BASE_URL: z.url(),
});

/** A contract ID env var that is still the deployment placeholder is treated as "not yet configured". */
function resolveContractId(value: string | undefined): string | undefined {
  return value === undefined || value === CONTRACT_ID_PLACEHOLDER || value.length === 0 ? undefined : value;
}

export interface PublicEnv {
  readonly appName: string;
  readonly network: NetworkName;
  readonly networkConfig: NetworkConfig;
  readonly registryContractId?: string;
  readonly orderContractId?: string;
  readonly apiBaseUrl: string;
}

let cachedEnv: PublicEnv | undefined;

/**
 * Parses and validates `NEXT_PUBLIC_*` configuration. Throws
 * {@link ConfigurationError} on missing/invalid values rather than falling
 * back to a default, so the application fails fast instead of running
 * against a silently wrong network.
 */
export function getPublicEnv(): PublicEnv {
  if (cachedEnv) {
    return cachedEnv;
  }

  const parsed = rawEnvSchema.safeParse({
    NEXT_PUBLIC_APP_NAME: process.env.NEXT_PUBLIC_APP_NAME,
    NEXT_PUBLIC_NETWORK: process.env.NEXT_PUBLIC_NETWORK,
    NEXT_PUBLIC_NETWORK_PASSPHRASE: process.env.NEXT_PUBLIC_NETWORK_PASSPHRASE,
    NEXT_PUBLIC_SOROBAN_RPC_URL: process.env.NEXT_PUBLIC_SOROBAN_RPC_URL,
    NEXT_PUBLIC_HORIZON_URL: process.env.NEXT_PUBLIC_HORIZON_URL,
    NEXT_PUBLIC_REGISTRY_CONTRACT_ID: process.env.NEXT_PUBLIC_REGISTRY_CONTRACT_ID,
    NEXT_PUBLIC_ORDER_CONTRACT_ID: process.env.NEXT_PUBLIC_ORDER_CONTRACT_ID,
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
  });

  if (!parsed.success) {
    throw new ConfigurationError(
      parsed.error.issues.map((issue) => `${issue.path.join(".") || "env"}: ${issue.message}`).join("; "),
    );
  }

  const data = parsed.data;
  const networkConfig = resolveNetworkConfig(data.NEXT_PUBLIC_NETWORK, {
    rpcUrl: data.NEXT_PUBLIC_SOROBAN_RPC_URL,
    horizonUrl: data.NEXT_PUBLIC_HORIZON_URL,
  });

  // Guards against the classic Stellar footgun: a passphrase that doesn't
  // match the selected network would let the wallet sign against the wrong
  // chain while the UI still claims the configured network (see spec §11).
  if (networkConfig.networkPassphrase !== data.NEXT_PUBLIC_NETWORK_PASSPHRASE) {
    throw new ConfigurationError(
      `NEXT_PUBLIC_NETWORK_PASSPHRASE does not match the passphrase for NEXT_PUBLIC_NETWORK="${data.NEXT_PUBLIC_NETWORK}". ` +
        `Expected "${networkConfig.networkPassphrase}", got "${data.NEXT_PUBLIC_NETWORK_PASSPHRASE}".`,
    );
  }

  const registryContractId = resolveContractId(data.NEXT_PUBLIC_REGISTRY_CONTRACT_ID);
  const orderContractId = resolveContractId(data.NEXT_PUBLIC_ORDER_CONTRACT_ID);

  cachedEnv = {
    appName: data.NEXT_PUBLIC_APP_NAME,
    network: data.NEXT_PUBLIC_NETWORK,
    networkConfig,
    ...(registryContractId !== undefined && { registryContractId }),
    ...(orderContractId !== undefined && { orderContractId }),
    apiBaseUrl: data.NEXT_PUBLIC_API_BASE_URL,
  };

  return cachedEnv;
}
