import type { NextConfig } from "next";

const withPWA = require("next-pwa")({
  dest: "public",
  register: true,
  skipWaiting: true,
  disable: process.env.NODE_ENV === "production", //production
  runtimeCaching: [
    {
      urlPattern: /\.(?:css|js)$/,
      handler: "NetworkFirst",
      options: {
        cacheName: "static-resources",
        expiration: {
          maxEntries: 64,
          maxAgeSeconds: 24 * 60 * 60, // 24 jam fallback offline
        },
      },
    },
  ],
});

const nextConfig: NextConfig = {
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
    "pug-widow-rewind.ngrok-free.dev",
    "*.ngrok-free.dev",
    "*.ngrok.io",
  ],
  images: {
    unoptimized: true,
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
      {
        protocol: "https",
        hostname: "*.ngrok-free.dev",
        pathname: "/**",
      },
      {
        protocol: "https",
        hostname: "*.ngrok.io",
        pathname: "/**",
      },
    ],
  },
  async headers() {
    return [
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

export default withPWA(nextConfig);
