import type { Metadata, Viewport } from "next";
import { SerwistProvider } from "@serwist/turbopack/react";
import "./globals.css";
import { Providers } from "@/components/providers";
import { THEME_INIT_SCRIPT } from "@/lib/theme";

export const metadata: Metadata = {
  title: {
    default: "I-GMP",
    template: "%s | I-GMP",
  },
  applicationName: "I-GMP",
  description: "Sistem inspeksi dan monitoring Good Manufacturing Practices.",
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
    title: "I-GMP",
  },
  formatDetection: {
    telephone: false,
    email: false,
    address: false,
  },
  robots: {
    index: false,
    follow: false,
  },
  icons: {
    icon: [{ url: "/Logo_plant_New.png", type: "image/png" }],
    shortcut: [{ url: "/Logo_plant_New.png", type: "image/png" }],
    apple: [{ url: "/Logo_plant_New.png", type: "image/png" }],
  },
};

export const viewport: Viewport = {
  colorScheme: "light dark",
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#1b3a6f" },
    { media: "(prefers-color-scheme: dark)", color: "#0f1f3d" },
  ],
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  viewportFit: "cover",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
      </head>
      <body
        className="font-sans antialiased"
        suppressHydrationWarning
      >
        <SerwistProvider swUrl="/serwist/sw.js" disable={process.env.NODE_ENV === "development"}>
          <Providers>
            {children}
          </Providers>
        </SerwistProvider>
      </body>
    </html>
  );
}
