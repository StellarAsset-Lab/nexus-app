# @nexus/contracts-order

Generated TypeScript client bindings for the Nexus Order Soroban contract.

Generated from the contract repository's local build artifact (no Order contract has been deployed yet):

```bash
stellar contract bindings typescript \
  --wasm <path-to>/nexus_order.wasm \
  --output-dir packages/contracts/order \
  --overwrite
```

Once the Order contract is deployed (Phase 8), regenerate from the deployed contract ID instead:

```bash
stellar contract bindings typescript --contract-id <ORDER_CONTRACT_ID> --output-dir packages/contracts/order --overwrite
```

Do not hand-edit `src/index.ts`. It is generated and is the source of truth for the Order contract ABI; regenerate it with `pnpm bindings` (see `scripts/generate-bindings.ts`) whenever the contract changes.

## Use it

```ts
import { Client } from "@nexus/contracts-order";

const order = new Client({
  contractId: "<ORDER_CONTRACT_ID>",
  networkPassphrase: "Test SDF Network ; September 2015",
  rpcUrl: "https://soroban-testnet.stellar.org",
});

const { result } = await order.get_order({ order_id: 1n });
```

`packages/sdk` wraps this generated client with the application's own network configuration, error handling, and read/write helpers — most application code should go through the SDK rather than this package directly.
