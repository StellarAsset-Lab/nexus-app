import { Badge } from "@nexus/ui";
import { NetworkGlobe } from "@/components/home/network-globe";
import { getPublicEnv } from "@/lib/env";

function SectionHeading({ eyebrow, title }: { eyebrow: string; title: string }) {
  return (
    <div className="mb-4">
      <p className="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">{eyebrow}</p>
      <h2 className="mt-1 text-xl font-semibold text-slate-900 dark:text-slate-100">{title}</h2>
    </div>
  );
}

function PreviewCard({ title, description, href }: { title: string; description: string; href: string }) {
  return (
    <a
      href={href}
      className="block rounded-lg border border-slate-200 p-5 transition-colors hover:border-slate-400 dark:border-slate-800 dark:hover:border-slate-600"
    >
      <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">{title}</h3>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{description}</p>
    </a>
  );
}

export default function HomePage() {
  const env = getPublicEnv();

  return (
    <main>
      {/* 1-4: product statement, problem, how it works, network visualization */}
      <section className="relative overflow-hidden border-b border-slate-200 dark:border-slate-800">
        <div className="mx-auto grid max-w-7xl gap-10 px-4 py-16 md:grid-cols-2 md:items-center">
          <div>
            <Badge tone="info">Stellar Testnet</Badge>
            <h1 className="mt-4 text-3xl font-semibold tracking-tight text-slate-900 dark:text-slate-100 sm:text-4xl">
              A standardized integration surface for Stellar asset settlement
            </h1>
            <p className="mt-4 text-base text-slate-600 dark:text-slate-400">
              Financial applications that want to offer Stellar assets today have to build eligibility checks, order
              lifecycle handling, settlement, and reconciliation themselves, against a protocol layer that was not
              designed for that integration pattern. Nexus provides that layer: a Registry for qualifying assets,
              distributors, and buyers, and an Order gateway that settles atomically once both sides are funded.
            </p>
            <p className="mt-4 text-sm text-slate-500 dark:text-slate-400">
              Discover, qualify, order, settle, verify, and reconcile — through typed contract access, a Go API over
              indexed protocol state, and a real Testnet integration path.
            </p>
          </div>
          <NetworkGlobe />
        </div>
      </section>

      {/* 5: live Testnet proof path */}
      <section className="border-b border-slate-200 px-4 py-12 dark:border-slate-800">
        <div className="mx-auto max-w-7xl">
          <SectionHeading eyebrow="Try it" title="Live Testnet proof path" />
          <p className="max-w-2xl text-sm text-slate-600 dark:text-slate-400">
            The Distribution Sandbox walks through a real order lifecycle on Stellar Testnet — asset and distributor
            selection, eligibility check, order creation, payment and asset funding, and settlement — using a
            connected wallet and real transaction confirmations. It also has a clearly labeled Simulation mode for
            walkthroughs that should not write to Testnet.
          </p>
          <a
            href="/sandbox"
            className="mt-4 inline-block rounded-md border border-slate-900 px-4 py-2 text-sm font-medium text-slate-900 hover:bg-slate-900 hover:text-white dark:border-slate-100 dark:text-slate-100 dark:hover:bg-slate-100 dark:hover:text-slate-900"
          >
            Open the sandbox
          </a>
        </div>
      </section>

      {/* 6-9: asset explorer, transaction lifecycle, developer integration, operations previews */}
      <section className="border-b border-slate-200 px-4 py-12 dark:border-slate-800">
        <div className="mx-auto max-w-7xl">
          <SectionHeading eyebrow="Explore" title="Indexed protocol state" />
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <PreviewCard
              title="Asset Explorer"
              description="Registered assets, distributors, and eligibility authorities."
              href="/assets"
            />
            <PreviewCard
              title="Transaction Center"
              description="Order lifecycle state from creation through settlement, cancellation, or expiry."
              href="/transactions"
            />
            <PreviewCard
              title="Developer Portal"
              description="SDK quickstart, API reference, contract ABI, and event reference."
              href="/developers"
            />
            <PreviewCard
              title="Operations Console"
              description="Indexer checkpoint, ingest lag, reconciliation mismatches, and stale transactions."
              href="/operations"
            />
          </div>
        </div>
      </section>

      {/* 10: system status */}
      <section className="border-b border-slate-200 px-4 py-12 dark:border-slate-800">
        <div className="mx-auto max-w-7xl">
          <SectionHeading eyebrow="Reliability" title="System status" />
          <p className="max-w-2xl text-sm text-slate-600 dark:text-slate-400">
            Web, API, database, indexer, worker, Stellar RPC, and contract health, measured directly rather than
            assumed.
          </p>
          <a href="/status" className="mt-3 inline-block text-sm font-medium text-blue-600 dark:text-blue-400">
            View system status →
          </a>
        </div>
      </section>

      {/* 11: documentation */}
      <section className="px-4 py-12">
        <div className="mx-auto max-w-7xl">
          <SectionHeading eyebrow="Build" title="Documentation" />
          <div className="grid gap-4 sm:grid-cols-3">
            <PreviewCard
              title="Quickstart"
              description="Install the SDK, configure the network, read an asset, create and settle an order."
              href="/developers/quickstart"
            />
            <PreviewCard title="API reference" description="Every Go API endpoint, its parameters, and its error shape." href="/developers/api" />
            <PreviewCard
              title="Contract reference"
              description="The real Registry and Order ABI, generated from the deployed contracts."
              href="/developers/contracts"
            />
          </div>
          <p className="mt-8 text-xs text-slate-400 dark:text-slate-600">
            Network: {env.network} · Registry contract: {env.registryContractId ?? "not yet deployed"} · Order
            contract: {env.orderContractId ?? "not yet deployed"}
          </p>
        </div>
      </section>
    </main>
  );
}
