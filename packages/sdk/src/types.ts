export type NetworkName = "testnet" | "mainnet";

export interface NetworkConfig {
  readonly name: NetworkName;
  readonly networkPassphrase: string;
  readonly rpcUrl: string;
  readonly horizonUrl: string;
}
