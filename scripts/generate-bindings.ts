#!/usr/bin/env tsx
/**
 * Regenerates packages/contracts/{registry,order} from the Wasm artifacts
 * recorded in packages/contracts/provenance.json, using the real Stellar
 * CLI (`stellar contract bindings typescript`) — never hand-written.
 *
 * Run via `pnpm bindings`. Requires the `stellar` CLI on PATH and a local
 * checkout of https://github.com/StellarAsset-Lab/nexus-contract; point
 * NEXUS_CONTRACT_REPO at it if it isn't a sibling of this repository.
 *
 * This only regenerates from the recorded Wasm — it does not update
 * provenance.json. After running this against a new Wasm build, update the
 * `sha256`/`sourceCommit` fields in provenance.json by hand and re-run
 * `pnpm verify-contracts` to confirm the new artifact matches, per
 * docs/contract-integration.md.
 */
import { existsSync, readFileSync, rmSync } from "node:fs";
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

function requireStellarCli(): void {
  try {
    execFileSync("stellar", ["--version"], { stdio: "ignore" });
  } catch {
    console.error("✗ The `stellar` CLI is not on PATH. Install it: https://developers.stellar.org/docs/tools/developer-tools/cli/install-cli");
    process.exit(1);
  }
}

function main(): void {
  requireStellarCli();

  const manifestPath = resolve(REPO_ROOT, "packages/contracts/provenance.json");
  const provenance = JSON.parse(readFileSync(manifestPath, "utf-8")) as Provenance;

  if (!existsSync(CONTRACT_REPO)) {
    console.error(
      `✗ Contract repository not found at ${CONTRACT_REPO}.\n` +
        `  Clone ${provenance.sourceRepository} there, or set NEXUS_CONTRACT_REPO to its location.`,
    );
    process.exit(1);
  }

  for (const [name, contract] of Object.entries(provenance.contracts)) {
    const wasmPath = resolve(CONTRACT_REPO, WASM_RELATIVE_DIR, contract.wasmFile);
    if (!existsSync(wasmPath)) {
      console.error(`✗ ${name}: Wasm artifact not found at ${wasmPath}. Build the contract first (cargo build --release --target wasm32v1-none).`);
      process.exit(1);
    }

    const outputDir = resolve(REPO_ROOT, contract.outputDir);
    console.log(`Generating ${name} bindings from ${wasmPath} into ${outputDir}...`);
    execFileSync(
      "stellar",
      ["contract", "bindings", "typescript", "--wasm", wasmPath, "--output-dir", outputDir, "--overwrite"],
      { stdio: "inherit" },
    );

    // The CLI's --overwrite also replaces package.json, tsconfig.json, and
    // README.md with its own generic boilerplate, and adds its own
    // .gitignore. Only src/index.ts is the actual generated ABI this repo
    // treats as generated; the rest are hand-maintained project scaffolding
    // (see each package's README), so restore them from git and drop the
    // CLI's .gitignore.
    const restorePaths = ["package.json", "tsconfig.json", "README.md"].map((f) => resolve(outputDir, f));
    execFileSync("git", ["checkout", "--", ...restorePaths], { cwd: REPO_ROOT, stdio: "inherit" });
    rmSync(resolve(outputDir, ".gitignore"), { force: true });
  }

  console.log("\nDone. Run `pnpm verify-contracts` to confirm the regenerated bindings match provenance.json.");
}

main();
