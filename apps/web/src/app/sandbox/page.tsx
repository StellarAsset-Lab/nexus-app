"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { Result } from "@stellar/stellar-sdk/contract";
import { buildCreateOrder, buildFundAsset, buildFundPayment, buildCancelOrder, buildExpireOrder, buildSettle } from "@nexus/sdk";
import { Badge, Button, Card, CardBody, CardHeader, EmptyState, ErrorState, LoadingState, StatusPill } from "@nexus/ui";
import { useWallet } from "@/components/wallet/wallet-context";
import { TransactionFlowStatus } from "@/components/sandbox/transaction-flow-status";
import { useTransactionFlow } from "@/hooks/use-transaction-flow";
import { useOrder } from "@/hooks/use-transactions";
import { useNetwork } from "@/hooks/use-status";
import { getNexusClient } from "@/lib/nexus-client";
import { getPublicEnv } from "@/lib/env";
import { orderStatusTone } from "@/lib/status-tone";

const INPUT_CLASSES =
  "w-full rounded-md border border-slate-300 px-3 py-2 text-sm dark:border-slate-600 dark:bg-slate-900";
const LABEL_CLASSES = "block text-xs font-medium text-slate-600 dark:text-slate-400";

function OrderCreationForm() {
  const wallet = useWallet();
  const network = useNetwork();
  const { flow, run, reset } = useTransactionFlow<Result<bigint>>();
  const [form, setForm] = useState({
    distributor: "",
    buyer: "",
    asset: "",
    paymentAsset: "",
    assetAmount: "",
    paymentAmount: "",
    expiresAtLedger: "",
  });
  const [createdOrderId, setCreatedOrderId] = useState<bigint>();

  function update(field: keyof typeof form) {
    return (e: React.ChangeEvent<HTMLInputElement>) => setForm((prev) => ({ ...prev, [field]: e.target.value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    reset();
    setCreatedOrderId(undefined);
    const client = getNexusClient();
    const orderId = await run(() =>
      buildCreateOrder(client, {
        distributor: form.distributor.trim(),
        buyer: form.buyer.trim(),
        asset: form.asset.trim(),
        paymentAsset: form.paymentAsset.trim(),
        assetAmount: BigInt(form.assetAmount || "0"),
        paymentAmount: BigInt(form.paymentAmount || "0"),
        expiresAtLedger: Number(form.expiresAtLedger || "0"),
      }),
    );
    if (orderId !== undefined) {
      setCreatedOrderId(orderId);
    }
  }

  const signer = wallet.getSigner();

  return (
    <form onSubmit={(e) => void handleSubmit(e)} className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <div>
          <label className={LABEL_CLASSES}>Distributor address</label>
          <input required value={form.distributor} onChange={update("distributor")} placeholder="G…" className={INPUT_CLASSES} />
        </div>
        <div>
          <label className={LABEL_CLASSES}>Buyer address</label>
          <input required value={form.buyer} onChange={update("buyer")} placeholder="G…" className={INPUT_CLASSES} />
        </div>
        <div>
          <label className={LABEL_CLASSES}>Asset (contract address)</label>
          <input required value={form.asset} onChange={update("asset")} placeholder="C…" className={INPUT_CLASSES} />
        </div>
        <div>
          <label className={LABEL_CLASSES}>Payment asset (contract address)</label>
          <input required value={form.paymentAsset} onChange={update("paymentAsset")} placeholder="C…" className={INPUT_CLASSES} />
        </div>
        <div>
          <label className={LABEL_CLASSES}>Asset amount (exact integer, no decimals)</label>
          <input required inputMode="numeric" value={form.assetAmount} onChange={update("assetAmount")} placeholder="1000000" className={INPUT_CLASSES} />
        </div>
        <div>
          <label className={LABEL_CLASSES}>Payment amount (exact integer, no decimals)</label>
          <input required inputMode="numeric" value={form.paymentAmount} onChange={update("paymentAmount")} placeholder="5000000" className={INPUT_CLASSES} />
        </div>
        <div>
          <label className={LABEL_CLASSES}>Expires at ledger</label>
          <input required inputMode="numeric" value={form.expiresAtLedger} onChange={update("expiresAtLedger")} placeholder={network.data?.latestLedger !== undefined ? String(network.data.latestLedger + 500) : "e.g. current ledger + 500"} className={INPUT_CLASSES} />
        </div>
      </div>

      {!signer && (
        <p className="text-sm text-amber-600 dark:text-amber-400">
          Connect a wallet on the configured network (top right) before creating an order.
        </p>
      )}

      <div>
        <Button type="submit" disabled={!signer || flow.state !== "idle" && flow.state !== "failed" && flow.state !== "confirmed" && flow.state !== "rejected" && flow.state !== "expired"} loading={flow.state !== "idle" && flow.state !== "confirmed" && flow.state !== "failed" && flow.state !== "rejected" && flow.state !== "expired"}>
          Create order
        </Button>
      </div>

      <TransactionFlowStatus flow={flow} />

      {createdOrderId !== undefined && (
        <p className="text-sm text-emerald-700 dark:text-emerald-400">
          Order #{createdOrderId.toString()} created — manage it below.
        </p>
      )}
    </form>
  );
}

function OrderManagementPanel() {
  const [orderIdInput, setOrderIdInput] = useState("");
  const [lookupOrderId, setLookupOrderId] = useState<number>();
  const orderQuery = useOrder(lookupOrderId);
  const queryClient = useQueryClient();

  const fundPaymentFlow = useTransactionFlow<Result<void>>();
  const fundAssetFlow = useTransactionFlow<Result<void>>();
  const settleFlow = useTransactionFlow<Result<void>>();
  const cancelFlow = useTransactionFlow<Result<void>>();
  const expireFlow = useTransactionFlow<Result<void>>();

  function handleLookup(e: React.FormEvent) {
    e.preventDefault();
    const id = Number(orderIdInput);
    if (Number.isFinite(id) && id >= 0) {
      setLookupOrderId(id);
    }
  }

  function refreshOrder() {
    if (lookupOrderId !== undefined) {
      void queryClient.invalidateQueries({ queryKey: ["order", lookupOrderId] });
    }
  }

  async function runAction(flow: ReturnType<typeof useTransactionFlow<Result<void>>>, build: (orderId: bigint) => ReturnType<typeof buildFundPayment>) {
    if (lookupOrderId === undefined) return;
    await flow.run(() => build(BigInt(lookupOrderId)));
    refreshOrder();
  }

  return (
    <div className="flex flex-col gap-6">
      <form onSubmit={handleLookup} className="flex gap-2">
        <input
          value={orderIdInput}
          onChange={(e) => setOrderIdInput(e.target.value)}
          placeholder="Order ID"
          inputMode="numeric"
          className={INPUT_CLASSES + " max-w-xs"}
        />
        <Button type="submit" variant="secondary" disabled={orderIdInput.trim() === ""}>
          Look up
        </Button>
      </form>

      {lookupOrderId !== undefined && (
        <>
          {orderQuery.isPending && <LoadingState rows={2} label="Loading order…" />}
          {orderQuery.isError && (
            <ErrorState message="We could not load this order." detail={orderQuery.error.message} onRetry={() => orderQuery.refetch()} />
          )}
          {orderQuery.isSuccess && (
            <Card>
              <CardHeader className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Order #{orderQuery.data.orderId}</h3>
                <StatusPill tone={orderStatusTone(orderQuery.data.status)} label={orderQuery.data.status} />
              </CardHeader>
              <CardBody className="flex flex-col gap-4">
                <div className="grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
                  <div>
                    <p className="text-slate-500 dark:text-slate-400">Payment funded</p>
                    <Badge tone={orderQuery.data.paymentFunded ? "success" : "neutral"}>{orderQuery.data.paymentFunded ? "Yes" : "No"}</Badge>
                  </div>
                  <div>
                    <p className="text-slate-500 dark:text-slate-400">Asset funded</p>
                    <Badge tone={orderQuery.data.assetFunded ? "success" : "neutral"}>{orderQuery.data.assetFunded ? "Yes" : "No"}</Badge>
                  </div>
                  <div>
                    <p className="text-slate-500 dark:text-slate-400">Expires at ledger</p>
                    <p className="font-medium text-slate-900 dark:text-slate-100">{orderQuery.data.expiresAtLedger}</p>
                  </div>
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => void runAction(fundPaymentFlow, (id) => buildFundPayment(getNexusClient(), id))}
                  >
                    Fund payment
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => void runAction(fundAssetFlow, (id) => buildFundAsset(getNexusClient(), id))}
                  >
                    Fund asset
                  </Button>
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={() => void runAction(settleFlow, (id) => buildSettle(getNexusClient(), id))}
                  >
                    Settle
                  </Button>
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => void runAction(cancelFlow, (id) => buildCancelOrder(getNexusClient(), id))}
                  >
                    Cancel
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => void runAction(expireFlow, (id) => buildExpireOrder(getNexusClient(), id))}
                  >
                    Expire
                  </Button>
                </div>

                <TransactionFlowStatus flow={fundPaymentFlow.flow} />
                <TransactionFlowStatus flow={fundAssetFlow.flow} />
                <TransactionFlowStatus flow={settleFlow.flow} />
                <TransactionFlowStatus flow={cancelFlow.flow} />
                <TransactionFlowStatus flow={expireFlow.flow} />
              </CardBody>
            </Card>
          )}
        </>
      )}
    </div>
  );
}

export default function SandboxPage() {
  const env = getPublicEnv();

  if (env.orderContractId === undefined) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-10">
        <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">Sandbox</h1>
        <Card className="mt-6">
          <CardBody>
            <EmptyState message="The Order contract has not been deployed yet — the Sandbox will be functional once NEXT_PUBLIC_ORDER_CONTRACT_ID is configured." />
          </CardBody>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-10">
      <h1 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">Sandbox</h1>
      <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
        Create and drive real Order contract transactions on {env.network}, signed by a connected wallet — every
        transaction here is real and submitted to the network.
      </p>

      <h2 className="mt-8 text-lg font-semibold text-slate-900 dark:text-slate-100">Create Order</h2>
      <Card className="mt-3">
        <CardBody>
          <OrderCreationForm />
        </CardBody>
      </Card>

      <h2 className="mt-8 text-lg font-semibold text-slate-900 dark:text-slate-100">Manage Existing Order</h2>
      <Card className="mt-3">
        <CardBody>
          <OrderManagementPanel />
        </CardBody>
      </Card>
    </div>
  );
}
