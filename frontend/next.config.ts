import type { NextConfig } from "next";

const publicBaseUrl =
  process.env.S3_PUBLIC_BASE_URL || process.env.NEXT_PUBLIC_SUPABASE_HOSTNAME;

function getHostname(input: string | undefined): string | null {
  if (!input) return null;
  try {
    const urlString = String(input);
    const formattedUrl = urlString.startsWith("http")
      ? urlString
      : `https://${urlString}`;
    return new URL(formattedUrl).hostname;
  } catch {
    return null;
  }
}

const imageHostname = getHostname(publicBaseUrl);

const nextConfig: NextConfig = {
  images: {
    // *** Uncomment below section for LOCAL TEST ONLY. ***
    dangerouslyAllowLocalIP: true,

    // *** For Deployment (For local test, comment this section). ***
    remotePatterns: imageHostname
      ? [
          {
            protocol: "http",
            hostname: "localhost",
            port: "9000",
            pathname: "/**",
          },
          {
            protocol: "https",
            hostname: imageHostname,
            port: "",
            pathname: "/storage/v1/object/public/**",
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
