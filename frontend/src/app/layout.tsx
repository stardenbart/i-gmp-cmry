import type { Metadata, Viewport } from "next";
import { SerwistProvider } from "@serwist/turbopack/react";
import "./globals.css";
import { Providers } from "@/components/providers";

const geistSans = { variable: "font-sans" };
const geistMono = { variable: "font-mono" };

export const metadata: Metadata = {
  title: "I-GMP Website",
  description: "Sistem Monitoring I-GMP Website",
  manifest: "/manifest.json",
  icons: {
    icon: [{ url: "/Logo_plant_New.png", type: "image/png" }],
    shortcut: [{ url: "/Logo_plant_New.png", type: "image/png" }],
    apple: [{ url: "/Logo_plant_New.png", type: "image/png" }],
  },
};

export const viewport: Viewport = {
  themeColor: "#000000",
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
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
