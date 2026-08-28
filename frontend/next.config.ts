import type { NextConfig } from "next";
import { withSerwist } from "@serwist/turbopack";

// next-pwa was retired: it hooks into the webpack config, but Next.js 16
// defaults `next build`/`next dev` to Turbopack, which ignores webpack()
// entirely — so next-pwa's service worker was silently never generated.
// @serwist/turbopack replaces it with a Turbopack-native equivalent: the
// actual service worker is built by the Route Handler at
// app/serwist/[path]/route.ts, with its source in app/sw.ts.

const nextConfig: NextConfig = {
  output: "standalone",
  compress: true,
  poweredByHeader: false,
  reactStrictMode: false,
  reactCompiler: true,
  compiler: {
    removeConsole: process.env.NODE_ENV === "production" ? { exclude: ["error", "warn"] } : false,
  },
  experimental: {
    optimizePackageImports: [
      "lucide-react",
      "@tanstack/react-query",
      "framer-motion",
      "recharts",
      "date-fns",
      "sonner",
    ],
  },
  turbopack: {},
  allowedDevOrigins: [
    "localhost:3000",
    "localhost:9000",
    "localhost:8080",
    "http://127.0.0.1:3000",
    "http://127.0.0.1:9000",
    "http://127.0.0.1:8080",
  ],
  images: {
    remotePatterns: [
      {
        protocol: "http",
        hostname: "localhost",
        port: "9000",
        pathname: "/**",
      },
      {
        protocol: "http",
        hostname: "minio",
        port: "9000",
        pathname: "/**",
      },
    ],
  },
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-Frame-Options", value: "SAMEORIGIN" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          {
            key: "Permissions-Policy",
            value: "geolocation=(), microphone=(), payment=(), usb=()",
          },
        ],
      },
      {
        source: "/monitoring-audit-bucket/:path*",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
      {
        source: "/uploads/:path*",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
      {
        source: "/serwist/:path*",
        headers: [
          {
            key: "Cache-Control",
            value: "no-cache, no-store, must-revalidate",
          },
        ],
      },
    ];
  },
  async redirects() {
    return [
      {
        source: "/cimory/dashboard/:path*",
        destination: "/cimory/all/dashboard/:path*",
        permanent: false,
      },
    ];
  },
  async rewrites() {
    const isProd = process.env.NODE_ENV === "production"; //development
    const backendUrl = process.env.INTERNAL_BACKEND_URL || process.env.NEXT_PUBLIC_BACKEND_URL || (isProd ? "http://backend:8080" : "http://127.0.0.1:8080");
    const minioUrl = process.env.MINIO_ENDPOINT
      ? (process.env.MINIO_ENDPOINT.startsWith("http") ? process.env.MINIO_ENDPOINT : `http://${process.env.MINIO_ENDPOINT}`)
      : (isProd ? "http://minio:9000" : "http://127.0.0.1:9000");

    return [
      {
        source: "/api/v1/:path*",
        destination: `${backendUrl}/api/v1/:path*`,
      },
      {
        source: "/monitoring-audit-bucket/:path*",
        destination: `${minioUrl}/monitoring-audit-bucket/:path*`,
      },
      {
        source: "/uploads/:path*",
        destination: `${backendUrl}/uploads/:path*`,
      },
    ];
  },
};

export default withSerwist(nextConfig);
