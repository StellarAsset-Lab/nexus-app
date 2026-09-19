# Contract Integration

Nexus integrates with two real Soroban contracts from
[`StellarAsset-Lab/nexus-contract`](https://github.com/StellarAsset-Lab/nexus-contract): **Registry** and **Order**.
Neither contract's source lives in this repository — only their generated TypeScript client bindings
(`packages/contracts/registry`, `packages/contracts/order`) and this application's integration code do.

## Provenance

`packages/contracts/provenance.json` records, for each contract, the exact Wasm artifact the checked-in bindings
were generated from: its SHA-256 hash and the `nexus-contract` commit it was built at. This is the only place that
claim is made — `pnpm verify-contracts` (see below) checks it against a real local build rather than trusting the
file blindly.

```json
{
  "contracts": {
    "registry": { "sha256": "...", "sourceCommit": "..." },
    "order": { "sha256": "...", "sourceCommit": "..." }
  }
}
```

## Regenerating bindings

Do not hand-edit `packages/contracts/*/src/index.ts` — it is generated and is the source of truth for each
contract's ABI. To regenerate:

```bash
git clone https://github.com/StellarAsset-Lab/nexus-contract ../nexus-contract   # if not already present
cd ../nexus-contract && cargo build --release --target wasm32v1-none
cd ../nexus-app
pnpm bindings          # scripts/generate-bindings.ts
pnpm verify-contracts  # scripts/verify-contracts.ts
```

`pnpm bindings` shells out to the real `stellar contract bindings typescript` CLI for both contracts, then restores
each package's hand-maintained `package.json`/`tsconfig.json`/`README.md` (the CLI's `--overwrite` replaces those
with generic boilerplate, which this repository does not use). After regenerating from a new Wasm build, update
`provenance.json`'s `sha256`/`sourceCommit` by hand and re-run `pnpm verify-contracts` to confirm.

If `NEXUS_CONTRACT_REPO` doesn't point at a sibling `../nexus-contract` checkout, set it explicitly:

```bash
NEXUS_CONTRACT_REPO=/path/to/nexus-contract pnpm bindings
```

## Registry contract

**Write methods**: `initialize`, `register_asset`, `deactivate_asset`, `set_eligibility`, `propose_admin`,
`accept_admin`, `cancel_admin_proposal`.

**Read methods**: `get_asset`, `is_asset_active`, `get_distribution`, `is_distribution_active`, `get_eligibility`,
`is_eligible`, `admin`, `pending_admin`.

**Events** (topic 0 is always the event name; see `internal/indexing/registry_events.go`):
`asset_registered`, `asset_deactivated`, `distribution_registered`, `distribution_revoked`, `eligibility_set`,
`eligibility_revoked`, `admin_transfer_proposed`, `admin_transferred`, `admin_transfer_cancelled`.

**Errors**: `Unauthorized`, `NotInitialized`, `AlreadyInitialized`, `PendingAdminAlreadySet`, `NoPendingAdmin`,
`AssetNotFound`, `AssetAlreadyActive`, `IssuerMismatch`, `DistributionNotFound`, `DistributionAlreadyActive`,
`EligibilityAuthorityMismatch`, `EligibilityNotFound`, `InvalidEligibilityExpiry`, `AssetInactive`,
`DistributionInactive`.

## Order contract

**Write methods**: `initialize`, `create_order`, `fund_payment`, `fund_asset`, `settle`, `cancel_order`,
`expire_order`, `pause`, `unpause`, `propose_admin`, `accept_admin`, `cancel_admin_proposal`.

**Read methods**: `get_order`, `order_exists`, `next_order_id`, `is_paused`, `admin`, `pending_admin`, `registry`.

**Events** (see `internal/indexing/order_events.go`): `order_created`, `payment_funded`, `asset_funded`,
`order_settled`, `order_cancelled`, `order_expired`, `gateway_paused`, `gateway_unpaused`,
`admin_transfer_proposed`, `admin_transferred`, `admin_transfer_cancelled`.

**Errors**: `Unauthorized`, `NotInitialized`, `AlreadyInitialized`, `PendingAdminAlreadySet`, `NoPendingAdmin`,
`Paused`, `AssetInactive`, `DistributionInactive`, `BuyerNotEligible`, `InvalidAmount`, `InvalidExpiry`,
`OrderNotFound`, `InvalidOrderStatus`, `PaymentAlreadyFunded`, `AssetAlreadyFunded`, `OrderExpired`, `NotExpired`,
`NotCancellable`, `NothingToCancel`.

## A note on event data encoding

Soroban's `#[contractevent]` macro defaults every event's `data_format` to `Map`, regardless of field count —
including zero- and one-field events. `internal/indexing/registry_events.go`'s `dataFields()` always expects a
`Map<Symbol, Val>`; do not "optimize" this into a bare-value shortcut for single-field events. This was verified
against both the real `soroban-sdk-macros` source and the compiled ABI's `ScSpecEventV0.DataFormat` for every event
in both contracts.

## Deployment status

Neither contract is deployed by this repository. `NEXT_PUBLIC_REGISTRY_CONTRACT_ID` and
`NEXT_PUBLIC_ORDER_CONTRACT_ID` are unset until a real deployment happens; every SDK write helper
(`packages/sdk/src/orders.ts`, `registry.ts`) throws a `ConfigurationError` rather than silently no-op'ing if called
before then, and the frontend Sandbox shows an honest "not yet deployed" state instead of a broken form.
