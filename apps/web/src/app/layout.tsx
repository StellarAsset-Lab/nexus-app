import type { Metadata } from "next";
import { AppShell, MobileNav, type NavItem } from "@nexus/ui";
import { ConnectWallet } from "@/components/wallet/connect-wallet";
import { AppProviders } from "@/providers/app-providers";
import "./globals.css";

export const metadata: Metadata = {
  title: "Nexus",
  description: "Discover, qualify, order, settle, verify, and reconcile Stellar assets through a standardized integration surface.",
};

const NAV_ITEMS: readonly NavItem[] = [
  { href: "/", label: "Home" },
  { href: "/assets", label: "Assets" },
  { href: "/transactions", label: "Transactions" },
  { href: "/sandbox", label: "Sandbox" },
  { href: "/developers", label: "Developers" },
  { href: "/status", label: "Status" },
  { href: "/operations", label: "Operations" },
];

const FOOTER_LINKS = [
  { href: "https://github.com/StellarAsset-Lab/nexus-app", label: "GitHub" },
  { href: "https://github.com/StellarAsset-Lab/nexus-contract", label: "Contracts" },
];

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AppProviders>
          <AppShell
            appName="Nexus"
            navItems={NAV_ITEMS}
            footerLinks={FOOTER_LINKS}
            headerRightSlot={<ConnectWallet />}
            mobileNavSlot={<MobileNav navItems={NAV_ITEMS} />}
          >
            {children}
          </AppShell>
        </AppProviders>
      </body>
    </html>
  );
}
