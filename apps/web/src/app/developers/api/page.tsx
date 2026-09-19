import Link from "next/link";
import { Badge, Card, CardBody, CardHeader } from "@nexus/ui";

interface Endpoint {
  readonly method: string;
  readonly path: string;
  readonly summary: string;
  readonly params?: readonly { name: string; description: string }[];
  readonly errors?: readonly { code: string; status: number }[];
}

const ENDPOINTS: readonly Endpoint[] = [
  { method: "GET", path: "/healthz", summary: "Liveness probe. Always 200 if the process is running." },
  {
    method: "GET",
    path: "/readyz",
    summary: "Readiness probe — checks the database connection.",
    errors: [{ code: "NOT_READY", status: 503 }],
  },
  {
    method: "GET",
    path: "/api/v1/network",
    summary: "Current network config, live RPC health, and the indexer's checkpoint.",
  },
  {
    method: "GET",
    path: "/api/v1/assets",
    summary: "List assets the Registry has ever registered, keyset-paginated.",
    params: [
      { name: "active", description: "true/false — filter by current active status. Omit for both." },
      { name: "cursor", description: "The asset address from the previous page's nextCursor." },
      { name: "limit", description: "Page size, default 50, max 200." },
    ],
  },
  {
    method: "GET",
    path: "/api/v1/assets/{asset}",
    summary: "A single asset by its contract address.",
    errors: [{ code: "ASSET_NOT_FOUND", status: 404 }],
  },
  {
    method: "GET",
    path: "/api/v1/assets/{asset}/activity",
    summary: "Recent orders referencing this asset.",
    errors: [{ code: "ASSET_NOT_FOUND", status: 404 }],
  },
  {
    method: "GET",
    path: "/api/v1/orders",
    summary: "List orders, keyset-paginated by (createdAtLedger, orderId) descending.",
    params: [
      { name: "status", description: "Created | Settled | Cancelled | Expired" },
      { name: "asset", description: "Filter by asset contract address." },
      { name: "distributor", description: "Filter by distributor address." },
      { name: "buyer", description: "Filter by buyer address." },
      { name: "createdAfterLedger", description: "Inclusive lower bound on createdAtLedger." },
      { name: "createdBeforeLedger", description: "Inclusive upper bound on createdAtLedger." },
      { name: "cursor", description: "From the previous page's nextCursor." },
      { name: "limit", description: "Page size, default 50, max 200." },
    ],
  },
  {
    method: "GET",
    path: "/api/v1/orders/{id}",
    summary: "A single order by its numeric ID.",
    errors: [
      { code: "INVALID_ORDER_ID", status: 400 },
      { code: "ORDER_NOT_FOUND", status: 404 },
    ],
  },
  {
    method: "GET",
    path: "/api/v1/transactions",
    summary: "List transactions the worker has reconciled against RPC, newest first.",
    params: [
      { name: "status", description: "Pending | Confirmed | Failed" },
      { name: "cursor", description: "From the previous page's nextCursor." },
      { name: "limit", description: "Page size, default 50, max 200." },
    ],
  },
  {
    method: "GET",
    path: "/api/v1/transactions/{hash}",
    summary: "A single transaction's real, reconciled status.",
    errors: [{ code: "TRANSACTION_NOT_FOUND", status: 404 }],
  },
  {
    method: "GET",
    path: "/api/v1/events",
    summary: "List raw indexed contract events, keyset-paginated by (ledger, eventId) ascending.",
    params: [
      { name: "eventType", description: "e.g. AssetRegistered, OrderCreated — see the contract reference." },
      { name: "contractId", description: "Filter by emitting contract address." },
      { name: "transactionHash", description: "Every event from one transaction." },
      { name: "cursor", description: "From the previous page's nextCursor." },
      { name: "limit", description: "Page size, default 50, max 200." },
    ],
  },
  {
    method: "GET",
    path: "/api/v1/status",
    summary: "Live, directly-checked health of every dependency (database, RPC, indexer).",
  },
];

export default function ApiReferencePage() {
  return (
    <div className="mx-auto max-w-4xl px-4 py-10">
      <Link href="/developers" className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100">
        ← Developer Portal
      </Link>
      <h1 className="mt-2 text-2xl font-semibold text-slate-900 dark:text-slate-100">API Reference</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Every endpoint the Go API serves. All responses are JSON; every error uses the stable shape{" "}
        <code className="font-mono text-xs">{"{ error: { code, message } }"}</code>.
      </p>

      <Card className="mt-6">
        <CardHeader>
          <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">SDK error taxonomy</h2>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
            Every error the TypeScript SDK throws extends <code className="font-mono text-xs">NexusError</code> with a
            stable <code className="font-mono text-xs">code</code> — distinct from the Go API&rsquo;s own error codes above.
          </p>
        </CardHeader>
        <CardBody>
          <ul className="flex flex-col gap-1 text-sm">
            <li><code className="font-mono text-xs">CONFIGURATION_ERROR</code> — missing or inconsistent SDK configuration (e.g. a contract ID used before deployment).</li>
            <li><code className="font-mono text-xs">VALIDATION_ERROR</code> — caller input failed local validation before any network call.</li>
            <li><code className="font-mono text-xs">RPC_ERROR</code> — a Stellar RPC/Horizon request failed.</li>
            <li><code className="font-mono text-xs">SIMULATION_ERROR</code> — Soroban simulation of a transaction failed.</li>
            <li><code className="font-mono text-xs">WALLET_REJECTED_ERROR</code> — the connected wallet declined to sign.</li>
            <li><code className="font-mono text-xs">SUBMISSION_ERROR</code> — RPC rejected the signed transaction on submission.</li>
            <li><code className="font-mono text-xs">CONFIRMATION_TIMEOUT_ERROR</code> — submitted but no final status within the configured timeout; may still confirm later.</li>
            <li><code className="font-mono text-xs">CONTRACT_ERROR</code> — the contract itself returned a <code className="font-mono text-xs">Result::Err</code> (see the contract reference for each contract&rsquo;s error variants).</li>
            <li><code className="font-mono text-xs">INDEXER_LAG_ERROR</code> — the indexed view is known to be behind the chain.</li>
          </ul>
        </CardBody>
      </Card>

      <div className="mt-6 flex flex-col gap-4">
        {ENDPOINTS.map((endpoint) => (
          <Card key={`${endpoint.method} ${endpoint.path}`}>
            <CardHeader className="flex items-center gap-3">
              <Badge tone="info">{endpoint.method}</Badge>
              <code className="font-mono text-sm text-slate-900 dark:text-slate-100">{endpoint.path}</code>
            </CardHeader>
            <CardBody>
              <p className="text-sm text-slate-600 dark:text-slate-400">{endpoint.summary}</p>
              {endpoint.params && endpoint.params.length > 0 && (
                <div className="mt-3">
                  <p className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Query parameters</p>
                  <ul className="mt-1 flex flex-col gap-1">
                    {endpoint.params.map((p) => (
                      <li key={p.name} className="text-sm">
                        <code className="font-mono text-xs text-slate-900 dark:text-slate-100">{p.name}</code>{" "}
                        <span className="text-slate-500 dark:text-slate-400">— {p.description}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {endpoint.errors && endpoint.errors.length > 0 && (
                <div className="mt-3">
                  <p className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Errors</p>
                  <ul className="mt-1 flex flex-col gap-1">
                    {endpoint.errors.map((e) => (
                      <li key={e.code} className="text-sm">
                        <code className="font-mono text-xs text-slate-900 dark:text-slate-100">{e.code}</code>{" "}
                        <span className="text-slate-500 dark:text-slate-400">— HTTP {e.status}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </CardBody>
          </Card>
        ))}
      </div>
    </div>
  );
}
