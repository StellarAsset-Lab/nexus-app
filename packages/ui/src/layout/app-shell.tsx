import type { ReactNode } from "react";
import { Footer, type FooterLink } from "./footer";
import { Header, type NavItem } from "./header";

export interface AppShellProps {
  readonly appName: string;
  readonly navItems: readonly NavItem[];
  readonly footerLinks: readonly FooterLink[];
  readonly logo?: ReactNode;
  readonly headerRightSlot?: ReactNode;
  readonly mobileNavSlot?: ReactNode;
  readonly children: ReactNode;
}

export function AppShell({
  appName,
  navItems,
  footerLinks,
  logo,
  headerRightSlot,
  mobileNavSlot,
  children,
}: AppShellProps) {
  return (
    <div className="flex min-h-screen flex-col">
      <Header appName={appName} navItems={navItems} logo={logo} rightSlot={headerRightSlot} mobileNavSlot={mobileNavSlot} />
      <main className="flex-1">{children}</main>
      <Footer appName={appName} links={footerLinks} />
    </div>
  );
}
