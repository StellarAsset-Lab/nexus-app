import Link from "next/link";
import { Card, CardBody, CardHeader } from "@nexus/ui";
import { CodeBlock } from "@/components/developers/code-block";

const INSTALL = `pnpm add @nexus/sdk`;

const CONFIGURE = `import { createNexusClient } from "@nexus/sdk";

const client = createNexusClient({
  network: "testnet",
  registryContractId: process.env.NEXT_PUBLIC_REGISTRY_CONTRACT_ID,
  orderContractId: process.env.NEXT_PUBLIC_ORDER_CONTRACT_ID,
});`;

const READ_ASSET = `import { getAsset, isEligible } from "@nexus/sdk";

const asset = await getAsset(client, assetContractId);
// undefined if the Registry has never seen this asset — never fabricated.

const eligible = await isEligible(client, {
  asset: assetContractId,
  distributor: distributorAddress,
  buyer: buyerAddress,
});`;

const CREATE_ORDER = `import { buildCreateOrder, executeTransaction } from "@nexus/sdk";

const tx = await buildCreateOrder(client, {
  distributor: distributorAddress,
  buyer: buyerAddress,
  asset: assetContractId,
  paymentAsset: paymentAssetContractId,
  assetAmount: 1_000_000n,   // exact i128 amount — always a bigint, never a float
  paymentAmount: 5_000_000n,
  expiresAtLedger: currentLedger + 500,
});

const { hash, result: orderId } = await executeTransaction(tx, wallet.getSigner()!, {
  onStateChange: (state) => console.log(state),
  // idle -> simulating -> awaiting-wallet -> pending -> confirmed
});`;

const SETTLE_ORDER = `import { buildFundPayment, buildFundAsset, buildSettle, executeTransaction } from "@nexus/sdk";

await executeTransaction(await buildFundPayment(client, orderId), buyerSigner);
await executeTransaction(await buildFundAsset(client, orderId), distributorSigner);
await executeTransaction(await buildSettle(client, orderId), anyPartySigner);`;

function Step({ index, title, description, code }: { index: number; title: string; description: string; code: string }) {
  return (
    <Card>
      <CardHeader>
        <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
          {index}. {title}
        </h2>
        <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{description}</p>
      </CardHeader>
      <CardBody>
        <CodeBlock code={code} />
      </CardBody>
    </Card>
  );
}

export default function QuickstartPage() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-10">
      <Link href="/developers" className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100">
        ← Developer Portal
      </Link>
      <h1 className="mt-2 text-2xl font-semibold text-slate-900 dark:text-slate-100">Quickstart</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Install the SDK, configure the network, read an asset, and create and settle an order — every snippet below
        is real code that runs against this repository&rsquo;s actual SDK exports.
      </p>

      <div className="mt-6 flex flex-col gap-4">
        <Step index={1} title="Install the SDK" description="A single TypeScript package covering configuration, reads, writes, and the wallet transaction lifecycle." code={INSTALL} />
        <Step index={2} title="Configure the network" description="Contract IDs are optional — they're undefined until deployment, and every SDK write helper throws a clear ConfigurationError if called before then." code={CONFIGURE} />
        <Step index={3} title="Read an asset and check eligibility" description="Every read returns undefined/false honestly when the on-chain state says so — never a guessed default." code={READ_ASSET} />
        <Step index={4} title="Create an order" description="Build (and simulate) a transaction, then hand it to a connected wallet's signer through executeTransaction, which drives the full lifecycle." code={CREATE_ORDER} />
        <Step index={5} title="Fund and settle" description="Payment and asset funding are independent steps — either party can fund first — followed by settlement once both sides are funded." code={SETTLE_ORDER} />
      </div>

      <p className="mt-8 text-sm text-slate-500 dark:text-slate-400">
        Try the full flow interactively in the <Link href="/sandbox" className="text-blue-600 hover:underline dark:text-blue-400">Sandbox</Link>, or read the{" "}
        <Link href="/developers/contracts" className="text-blue-600 hover:underline dark:text-blue-400">contract reference</Link> for every method, event, and error.
      </p>
    </div>
  );
}
