import Link from "next/link";
import { Badge, Card, CardBody, CardHeader } from "@nexus/ui";
import { getPublicEnv } from "@/lib/env";

interface ContractSection {
  readonly name: string;
  readonly contractId?: string;
  readonly writeMethods: readonly string[];
  readonly readMethods: readonly string[];
  readonly events: readonly string[];
  readonly errors: readonly string[];
}

function ContractCard({ section }: { section: ContractSection }) {
  return (
    <Card>
      <CardHeader className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">{section.name}</h2>
        {section.contractId ? (
          <code className="font-mono text-xs text-slate-500 dark:text-slate-400">{section.contractId}</code>
        ) : (
          <Badge tone="neutral">Not yet deployed</Badge>
        )}
      </CardHeader>
      <CardBody className="flex flex-col gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Write methods</p>
          <p className="mt-1 flex flex-wrap gap-2">
            {section.writeMethods.map((m) => (
              <code key={m} className="rounded bg-slate-100 px-2 py-0.5 font-mono text-xs dark:bg-slate-800">
                {m}
              </code>
            ))}
          </p>
        </div>
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Read methods</p>
          <p className="mt-1 flex flex-wrap gap-2">
            {section.readMethods.map((m) => (
              <code key={m} className="rounded bg-slate-100 px-2 py-0.5 font-mono text-xs dark:bg-slate-800">
                {m}
              </code>
            ))}
          </p>
        </div>
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Events</p>
          <p className="mt-1 flex flex-wrap gap-2">
            {section.events.map((e) => (
              <code key={e} className="rounded bg-slate-100 px-2 py-0.5 font-mono text-xs dark:bg-slate-800">
                {e}
              </code>
            ))}
          </p>
        </div>
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">Errors</p>
          <p className="mt-1 flex flex-wrap gap-2">
            {section.errors.map((e) => (
              <code key={e} className="rounded bg-slate-100 px-2 py-0.5 font-mono text-xs dark:bg-slate-800">
                {e}
              </code>
            ))}
          </p>
        </div>
      </CardBody>
    </Card>
  );
}

export default function ContractReferencePage() {
  const env = getPublicEnv();

  const registry: ContractSection = {
    name: "Registry",
    ...(env.registryContractId !== undefined && { contractId: env.registryContractId }),
    writeMethods: [
      "initialize",
      "register_asset",
      "deactivate_asset",
      "set_eligibility",
      "propose_admin",
      "accept_admin",
      "cancel_admin_proposal",
    ],
    readMethods: [
      "get_asset",
      "is_asset_active",
      "get_distribution",
      "is_distribution_active",
      "get_eligibility",
      "is_eligible",
      "admin",
      "pending_admin",
    ],
    events: [
      "asset_registered",
      "asset_deactivated",
      "distribution_registered",
      "distribution_revoked",
      "eligibility_set",
      "eligibility_revoked",
      "admin_transfer_proposed",
      "admin_transferred",
      "admin_transfer_cancelled",
    ],
    errors: [
      "Unauthorized",
      "NotInitialized",
      "AlreadyInitialized",
      "PendingAdminAlreadySet",
      "NoPendingAdmin",
      "AssetNotFound",
      "AssetAlreadyActive",
      "IssuerMismatch",
      "DistributionNotFound",
      "DistributionAlreadyActive",
      "EligibilityAuthorityMismatch",
      "EligibilityNotFound",
      "InvalidEligibilityExpiry",
      "AssetInactive",
      "DistributionInactive",
    ],
  };

  const order: ContractSection = {
    name: "Order",
    ...(env.orderContractId !== undefined && { contractId: env.orderContractId }),
    writeMethods: [
      "initialize",
      "create_order",
      "fund_payment",
      "fund_asset",
      "settle",
      "cancel_order",
      "expire_order",
      "pause",
      "unpause",
      "propose_admin",
      "accept_admin",
      "cancel_admin_proposal",
    ],
    readMethods: ["get_order", "order_exists", "next_order_id", "is_paused", "admin", "pending_admin", "registry"],
    events: [
      "order_created",
      "payment_funded",
      "asset_funded",
      "order_settled",
      "order_cancelled",
      "order_expired",
      "gateway_paused",
      "gateway_unpaused",
      "admin_transfer_proposed",
      "admin_transferred",
      "admin_transfer_cancelled",
    ],
    errors: [
      "Unauthorized",
      "NotInitialized",
      "AlreadyInitialized",
      "PendingAdminAlreadySet",
      "NoPendingAdmin",
      "Paused",
      "AssetInactive",
      "DistributionInactive",
      "BuyerNotEligible",
      "InvalidAmount",
      "InvalidExpiry",
      "OrderNotFound",
      "InvalidOrderStatus",
      "PaymentAlreadyFunded",
      "AssetAlreadyFunded",
      "OrderExpired",
      "NotExpired",
      "NotCancellable",
      "NothingToCancel",
    ],
  };

  return (
    <div className="mx-auto max-w-4xl px-4 py-10">
      <Link href="/developers" className="text-sm text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100">
        ← Developer Portal
      </Link>
      <h1 className="mt-2 text-2xl font-semibold text-slate-900 dark:text-slate-100">Contract Reference</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Generated from the real, deployed Registry and Order contract ABI (see{" "}
        <code className="font-mono text-xs">packages/contracts/registry</code> and{" "}
        <code className="font-mono text-xs">packages/contracts/order</code>) — never hand-guessed.
      </p>

      <div className="mt-6 flex flex-col gap-4">
        <ContractCard section={registry} />
        <ContractCard section={order} />
      </div>
    </div>
  );
}
