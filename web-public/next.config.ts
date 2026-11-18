import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone", // small self-contained build for Docker
  cacheComponents: true,
  partialPrefetching: true,
};

export default nextConfig;
