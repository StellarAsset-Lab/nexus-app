import { createNexusClient, type NexusClient } from "@nexus/sdk";
import { getPublicEnv } from "./env";

let cachedClient: NexusClient | undefined;

/**
 * The application's single {@link NexusClient} instance, built from the
 * validated public env. `orderContract`/`registryContract` on it are
 * undefined until their contract IDs are configured post-deployment — every
 * caller (the Sandbox pages) must handle that, never assume a contract is
 * live.
 */
export function getNexusClient(): NexusClient {
  if (cachedClient) {
    return cachedClient;
  }
  const env = getPublicEnv();
  cachedClient = createNexusClient({
    network: env.networkConfig.name,
    rpcUrl: env.networkConfig.rpcUrl,
    horizonUrl: env.networkConfig.horizonUrl,
    ...(env.registryContractId !== undefined && { registryContractId: env.registryContractId }),
    ...(env.orderContractId !== undefined && { orderContractId: env.orderContractId }),
  });
  return cachedClient;
}
