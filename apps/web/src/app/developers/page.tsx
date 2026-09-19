import Link from "next/link";
import { Card, CardBody } from "@nexus/ui";

const SECTIONS = [
  {
    href: "/developers/quickstart",
    title: "Quickstart",
    description: "Install the SDK, configure the network, read an asset, and create and settle an order.",
  },
  {
    href: "/developers/api",
    title: "API reference",
    description: "Every Go API endpoint, its parameters, and its error shape.",
  },
  {
    href: "/developers/contracts",
    title: "Contract reference",
    description: "The real Registry and Order ABI — methods, events, and errors — generated from the deployed contracts.",
  },
];

export default function DevelopersPage() {
  return (
    <div className="mx-auto max-w-4xl px-4 py-10">
      <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">Developer Portal</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Everything needed to integrate with Nexus: the TypeScript SDK, the Go API, and the real Registry and Order
        contract ABI.
      </p>

      <div className="mt-6 grid gap-4 sm:grid-cols-3">
        {SECTIONS.map((section) => (
          <Link key={section.href} href={section.href}>
            <Card className="h-full transition-colors hover:border-slate-400 dark:hover:border-slate-600">
              <CardBody>
                <h2 className="text-sm font-semibold text-slate-900 dark:text-slate-100">{section.title}</h2>
                <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{section.description}</p>
              </CardBody>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
