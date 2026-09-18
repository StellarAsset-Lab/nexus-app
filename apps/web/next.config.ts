import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ["@nexus/sdk", "@nexus/ui", "@nexus/contracts-registry", "@nexus/contracts-order"],
};

export default nextConfig;
