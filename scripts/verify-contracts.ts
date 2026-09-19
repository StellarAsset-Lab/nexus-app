#!/usr/bin/env tsx
/**
 * Verifies that the generated bindings checked into packages/contracts/*
 * really were produced from the exact Wasm artifacts recorded in
 * packages/contracts/provenance.json — never trust that a binding matches
 * its claimed source without checking the hash.
 *
 * Run via `pnpm verify-contracts`. Requires a local checkout of
 * https://github.com/StellarAsset-Lab/nexus-contract; point
 * NEXUS_CONTRACT_REPO at it if it isn't a sibling of this repository.
 */
import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { execFileSync } from "node:child_process";

interface ContractProvenance {
  readonly wasmFile: string;
  readonly sha256: string;
  readonly sourceCommit: string;
  readonly outputDir: string;
}

interface Provenance {
  readonly sourceRepository: string;
  readonly contracts: Record<string, ContractProvenance>;
}

const REPO_ROOT = process.cwd();
const CONTRACT_REPO = resolve(process.env.NEXUS_CONTRACT_REPO ?? resolve(REPO_ROOT, "..", "nexus-contract"));
const WASM_RELATIVE_DIR = "target/wasm32v1-none/release";

function sha256File(path: string): string {
  return createHash("sha256").update(readFileSync(path)).digest("hex");
}

function isAncestorCommit(repoDir: string, ancestor: string, descendant: string): boolean {
  try {
    execFileSync("git", ["merge-base", "--is-ancestor", ancestor, descendant], { cwd: repoDir, stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

function currentHead(repoDir: string): string {
  return execFileSync("git", ["rev-parse", "HEAD"], { cwd: repoDir }).toString().trim();
}

function main(): void {
  const manifestPath = resolve(REPO_ROOT, "packages/contracts/provenance.json");
  const provenance = JSON.parse(readFileSync(manifestPath, "utf-8")) as Provenance;

  if (!existsSync(CONTRACT_REPO)) {
    console.error(
      `✗ Contract repository not found at ${CONTRACT_REPO}.\n` +
        `  Clone ${provenance.sourceRepository} there, or set NEXUS_CONTRACT_REPO to its location.`,
    );
    process.exit(1);
  }

  const head = currentHead(CONTRACT_REPO);
  let failures = 0;

  for (const [name, contract] of Object.entries(provenance.contracts)) {
    const wasmPath = resolve(CONTRACT_REPO, WASM_RELATIVE_DIR, contract.wasmFile);
    if (!existsSync(wasmPath)) {
      console.error(`✗ ${name}: Wasm artifact not found at ${wasmPath}. Build the contract first (cargo build --release --target wasm32v1-none).`);
      failures++;
      continue;
    }

    const actualSha256 = sha256File(wasmPath);
    if (actualSha256 !== contract.sha256) {
      console.error(
        `✗ ${name}: Wasm hash mismatch.\n` +
          `  expected: ${contract.sha256}\n` +
          `  actual:   ${actualSha256}\n` +
          `  The generated bindings in ${contract.outputDir} no longer match this Wasm. Regenerate with \`pnpm bindings\` and update provenance.json.`,
      );
      failures++;
      continue;
    }

    if (!isAncestorCommit(CONTRACT_REPO, contract.sourceCommit, head)) {
      console.error(
        `✗ ${name}: recorded source commit ${contract.sourceCommit} is not an ancestor of the contract repo's current HEAD (${head}).\n` +
          `  The provenance record may be stale, or the local checkout is on an unrelated branch.`,
      );
      failures++;
      continue;
    }

    console.log(`✓ ${name}: Wasm hash and source commit verified (${contract.sourceCommit.slice(0, 12)}).`);
  }

  if (failures > 0) {
    console.error(`\n${failures} contract(s) failed verification.`);
    process.exit(1);
  }
}

main();
