import { AlbedoModule } from "@creit.tech/stellar-wallets-kit/modules/albedo";
import { FreighterModule } from "@creit.tech/stellar-wallets-kit/modules/freighter";
import { HanaModule } from "@creit.tech/stellar-wallets-kit/modules/hana";
import { LobstrModule } from "@creit.tech/stellar-wallets-kit/modules/lobstr";
import { RabetModule } from "@creit.tech/stellar-wallets-kit/modules/rabet";
import { xBullModule } from "@creit.tech/stellar-wallets-kit/modules/xbull";
import type { ModuleInterface } from "@creit.tech/stellar-wallets-kit";

/**
 * Wallet modules that need no external service credentials (no WalletConnect
 * project ID, no hardware/USB permissions). WalletConnect and hardware-wallet
 * modules can be added later once those credentials are actually configured
 * — they are deliberately omitted here rather than wired up with placeholder
 * values.
 */
export function createWalletModules(): ModuleInterface[] {
  return [
    new FreighterModule(),
    new xBullModule(),
    new LobstrModule(),
    new RabetModule(),
    new HanaModule(),
    new AlbedoModule(),
  ];
}
