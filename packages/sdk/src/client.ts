import { Server } from "@stellar/stellar-sdk/rpc";
import { Client as RegistryClient } from "@nexus/contracts-registry";
import { Client as OrderClient } from "@nexus/contracts-order";
import { createNexusConfig, type NexusConfigInput, type NexusConfig } from "./config";

export interface NexusClient {
  readonly config: NexusConfig;
  readonly rpc: Server;
  /** Undefined until `registryContractId` is configured (i.e. before deployment). */
  readonly registryContract?: RegistryClient;
  /** Undefined until `orderContractId` is configured (i.e. before deployment). */
  readonly orderContract?: OrderClient;
}

/**
 * Constructs the low-level Nexus client: validated network configuration,
 * an RPC connection, and generated contract clients for whichever contracts
 * are configured. Higher-level read/write helpers (see registry.ts,
 * orders.ts, transactions.ts) are built on top of this.
 */
export function createNexusClient(input: NexusConfigInput): NexusClient {
  const config = createNexusConfig(input);

  const allowHttp = config.network.rpcUrl.startsWith("http://");

  const rpc = new Server(config.network.rpcUrl, { allowHttp });

  const registryContract =
    config.registryContractId !== undefined
      ? new RegistryClient({
          contractId: config.registryContractId,
          networkPassphrase: config.network.networkPassphrase,
          rpcUrl: config.network.rpcUrl,
          allowHttp,
        })
      : undefined;

  const orderContract =
    config.orderContractId !== undefined
      ? new OrderClient({
          contractId: config.orderContractId,
          networkPassphrase: config.network.networkPassphrase,
          rpcUrl: config.network.rpcUrl,
          allowHttp,
        })
      : undefined;

  return {
    config,
    rpc,
    ...(registryContract !== undefined && { registryContract }),
    ...(orderContract !== undefined && { orderContract }),
  };
}
