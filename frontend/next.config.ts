import type { NextConfig } from "next";

const getHostname = (url?: string) => {
  if (!url) return null;
  try {
    return new URL(url).hostname;
  } catch {
    return null;
  }
};

const targetHost = getHostname(process.env.S3_PUBLIC_BASE_URL);

const nextConfig: NextConfig = {
  images: {
    // Safety net: serve every <Image> src as-is (no /_next/image, so no Vercel
    // transformations). User-uploaded and presigned images also set
    // `unoptimized` per component, and the remaining static assets are SVGs
    // (never optimized), so removing this flag currently changes nothing.
    // To optimize an image again, drop `unoptimized` from that component only,
    // use stable (non-presigned) URLs, and set minimumCacheTTL (e.g. 2678400).
    unoptimized: true,

    // Local MinIO fallback serves presigned URLs from localhost/private IPs;
    // Next's Image Optimizer blocks those as a potential SSRF vector unless
    // this is set. Only enabled when no public S3 base URL is configured.
    dangerouslyAllowLocalIP: !targetHost,

    remotePatterns: targetHost
      ? [
          {
            protocol: "https" as const,
            hostname: targetHost,
            port: "",
            pathname: "/**",
          },
          {
            protocol: "https" as const,
            hostname: "*.supabase.co",
            port: "",
            pathname: "/**",
          },
        ]
      : [
          {
            protocol: "http",
            hostname: "localhost",
            port: "9000",
            pathname: "/**",
          },
        ],
  },
};

export default nextConfig;
