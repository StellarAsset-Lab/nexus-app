import { z } from "zod";
import { ConfigurationError } from "./errors";
import { isNetworkName, resolveNetworkConfig } from "./network";
import type { NetworkConfig } from "./types";

const nexusConfigInputSchema = z.object({
  network: z.string().refine(isNetworkName, {
    message: 'network must be "testnet" or "mainnet"',
  }),
  rpcUrl: z.url().optional(),
  horizonUrl: z.url().optional(),
  /** Undefined until the Registry contract has been deployed and configured. */
  registryContractId: z.string().min(1).optional(),
  /** Undefined until the Order contract has been deployed and configured. */
  orderContractId: z.string().min(1).optional(),
});

export type NexusConfigInput = z.input<typeof nexusConfigInputSchema>;

export interface NexusConfig {
  readonly network: NetworkConfig;
  readonly registryContractId?: string;
  readonly orderContractId?: string;
}

/**
 * Validates raw SDK configuration and resolves it into a concrete
 * {@link NexusConfig}. Throws {@link ConfigurationError} rather than
 * returning null/undefined on invalid input, so callers cannot accidentally
 * proceed with a half-configured client.
 */
export function createNexusConfig(input: NexusConfigInput): NexusConfig {
  const parsed = nexusConfigInputSchema.safeParse(input);
  if (!parsed.success) {
    throw new ConfigurationError(
      parsed.error.issues.map((issue) => `${issue.path.join(".") || "config"}: ${issue.message}`).join("; "),
    );
  }

  const network = resolveNetworkConfig(parsed.data.network, {
    ...(parsed.data.rpcUrl !== undefined && { rpcUrl: parsed.data.rpcUrl }),
    ...(parsed.data.horizonUrl !== undefined && { horizonUrl: parsed.data.horizonUrl }),
  });

  return {
    network,
    ...(parsed.data.registryContractId !== undefined && { registryContractId: parsed.data.registryContractId }),
    ...(parsed.data.orderContractId !== undefined && { orderContractId: parsed.data.orderContractId }),
  };
}
