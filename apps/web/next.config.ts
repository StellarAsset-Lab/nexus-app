import { config as loadEnv } from "dotenv";
import { resolve } from "node:path";
import type { NextConfig } from "next";

// This monorepo keeps a single .env at the repo root (shared with the Go
// services), but Next.js only auto-loads .env files from the app's own
// directory (apps/web/). Load the root one explicitly before Next reads
// process.env for NEXT_PUBLIC_* inlining. Real environment variables already
// set (e.g. by a deployment platform) take precedence over the file.
loadEnv({ path: resolve(__dirname, "../../.env") });

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ["@nexus/sdk", "@nexus/ui", "@nexus/contracts-registry", "@nexus/contracts-order"],
  // Do not auto-generate AGENTS.md/CLAUDE.md — this project keeps no AI/agent
  // identity or tooling artifacts anywhere in the repository.
  agentRules: false,
};

export default nextConfig;
