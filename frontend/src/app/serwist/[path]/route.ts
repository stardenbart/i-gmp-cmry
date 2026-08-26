import { createSerwistRoute } from "@serwist/turbopack";

// Builds and serves the service worker (compiled from src/app/sw.ts) at
// request time via esbuild — this is what replaced next-pwa's build-time
// webpack emission, which Turbopack no longer executes.
export const { dynamic, dynamicParams, revalidate, generateStaticParams, GET } = createSerwistRoute({
  swSrc: "src/app/sw.ts",
  useNativeEsbuild: true,
});
