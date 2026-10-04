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
    // *** Uncomment below line for LOCAL TEST ONLY. ***
    // dangerouslyAllowLocalIP: true,

    remotePatterns: targetHost
      ? [
          ...(targetHost
            ? [
                {
                  protocol: "https" as const,
                  hostname: targetHost,
                  port: "",
                  pathname: "/**",
                },
              ]
            : []),
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
