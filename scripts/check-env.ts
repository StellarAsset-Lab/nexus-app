#!/usr/bin/env tsx
/**
 * Validates a local `.env` file against the full set of variables the
 * application stack requires (browser-safe and server-only). Run via
 * `pnpm check-env` (see Makefile / docs/local-development.md).
 *
 * This intentionally does not import apps/web's env module: that module
 * validates only the browser-safe subset at Next.js runtime, while this
 * script is a repo-wide preflight check covering the Go services' variables
 * too, and is meant to be run before any service starts.
 */
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { parse as parseDotenv } from "dotenv";
import { isNetworkName } from "@nexus/sdk";
import { z } from "zod";

const CONTRACT_ID_PLACEHOLDER = "<SET_AFTER_DEPLOYMENT>";
const START_LEDGER_PLACEHOLDER = "<OPTIONAL_START_LEDGER>";
const API_KEY_PEPPER_PLACEHOLDER = "<RANDOM_SERVER_SECRET>";

const envSchema = z.object({
  // Browser-safe
  NEXT_PUBLIC_APP_NAME: z.string().min(1),
  NEXT_PUBLIC_NETWORK: z.string().refine(isNetworkName, {
    message: 'must be "testnet" or "mainnet"',
  }),
  NEXT_PUBLIC_NETWORK_PASSPHRASE: z.string().min(1),
  NEXT_PUBLIC_SOROBAN_RPC_URL: z.url(),
  NEXT_PUBLIC_HORIZON_URL: z.url(),
  NEXT_PUBLIC_REGISTRY_CONTRACT_ID: z.string().min(1),
  NEXT_PUBLIC_ORDER_CONTRACT_ID: z.string().min(1),
  NEXT_PUBLIC_API_BASE_URL: z.url(),

  // Server-only
  DATABASE_URL: z.string().regex(/^postgres(?:ql)?:\/\//, 'must be a "postgres://" connection string'),
  API_BIND_ADDR: z.string().regex(/^([^:]+)?:\d+$/, 'must look like ":8080" or "host:8080"'),
  INDEXER_BIND_ADDR: z.string().regex(/^([^:]+)?:\d+$/, 'must look like ":8081" or "host:8081"'),
  WORKER_RECONCILE_INTERVAL: z.string().regex(/^\d+$/, "must be a positive integer number of seconds"),
  INDEXER_START_LEDGER: z.string().min(1),
  INDEXER_BATCH_LIMIT: z.string().regex(/^\d+$/, "must be a positive integer"),
  INDEXER_CONFIRMATION_LOOKBACK: z.string().regex(/^\d+$/, "must be a positive integer"),
  API_KEY_PEPPER: z.string().min(1),
});

function loadEnvFile(path: string): Record<string, string | undefined> {
  if (!existsSync(path)) {
    console.error(`✗ ${path} does not exist. Copy .env.example to .env and fill in real values first.`);
    process.exit(1);
  }
  return parseDotenv(readFileSync(path));
}

function main(): void {
  const envPath = resolve(process.cwd(), process.argv[2] ?? ".env");
  const env = loadEnvFile(envPath);

  const parsed = envSchema.safeParse(env);
  const warnings: string[] = [];

  if (env.NEXT_PUBLIC_REGISTRY_CONTRACT_ID === CONTRACT_ID_PLACEHOLDER) {
    warnings.push("NEXT_PUBLIC_REGISTRY_CONTRACT_ID is still the deployment placeholder.");
  }
  if (env.NEXT_PUBLIC_ORDER_CONTRACT_ID === CONTRACT_ID_PLACEHOLDER) {
    warnings.push("NEXT_PUBLIC_ORDER_CONTRACT_ID is still the deployment placeholder.");
  }
  if (env.INDEXER_START_LEDGER === START_LEDGER_PLACEHOLDER) {
    warnings.push("INDEXER_START_LEDGER is still the optional placeholder; the indexer will start from the persisted checkpoint or chain head.");
  }
  if (env.API_KEY_PEPPER === API_KEY_PEPPER_PLACEHOLDER) {
    warnings.push("API_KEY_PEPPER is still the example placeholder; replace it with a real random secret before enabling API key auth.");
  }

  if (!parsed.success) {
    console.error(`✗ ${envPath} is invalid:\n`);
    for (const issue of parsed.error.issues) {
      console.error(`  - ${issue.path.join(".") || "(root)"}: ${issue.message}`);
    }
    process.exit(1);
  }

  console.log(`✓ ${envPath} has all required variables in a valid shape.`);
  for (const warning of warnings) {
    console.warn(`  ⚠ ${warning}`);
  }
}

main();
