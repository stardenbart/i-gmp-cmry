import type { Metadata, Viewport } from "next";
import "./globals.css";
import { Providers } from "@/components/providers";

const geistSans = { variable: "font-sans" };
const geistMono = { variable: "font-mono" };

export const metadata: Metadata = {
  title: "Audit Monitoring System",
  description: "Sistem Monitoring Audit Internal",
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
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  );
}
