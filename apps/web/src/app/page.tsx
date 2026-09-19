import { ConnectWallet } from "@/components/wallet/connect-wallet";

export default function HomePage() {
  return (
    <main>
      <div className="flex justify-end p-4">
        <ConnectWallet />
      </div>
      <p>Nexus application scaffold. The product homepage lands in a later commit.</p>
    </main>
  );
}
