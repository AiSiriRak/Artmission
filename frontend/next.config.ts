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
