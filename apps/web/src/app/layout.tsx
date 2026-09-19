import type { Metadata } from "next";
import { AppProviders } from "@/providers/app-providers";
import "./globals.css";

export const metadata: Metadata = {
  title: "Nexus",
  description: "Discover, qualify, order, settle, verify, and reconcile Stellar assets through a standardized integration surface.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
