import { Buffer } from "buffer";
import { Address } from "@stellar/stellar-sdk";
import {
  AssembledTransaction,
  Client as ContractClient,
  ClientOptions as ContractClientOptions,
  MethodOptions,
  Result,
  Spec as ContractSpec,
} from "@stellar/stellar-sdk/contract";
import type {
  u32,
  i32,
  u64,
  i64,
  u128,
  i128,
  u256,
  i256,
  Option,
  Timepoint,
  Duration,
} from "@stellar/stellar-sdk/contract";
export * from "@stellar/stellar-sdk";
export * as contract from "@stellar/stellar-sdk/contract";
export * as rpc from "@stellar/stellar-sdk/rpc";

if (typeof window !== "undefined") {
  //@ts-ignore Buffer exists
  window.Buffer = window.Buffer || Buffer;
}





export interface EligibilityRecord {
  asset: string;
  buyer: string;
  distributor: string;
  valid_until_ledger: u32;
}

export const ContractError = {
  1: {message:"Unauthorized"},
  2: {message:"NotInitialized"},
  3: {message:"AlreadyInitialized"},
  4: {message:"PendingAdminAlreadySet"},
  5: {message:"NoPendingAdmin"},
  6: {message:"Paused"},
  7: {message:"AssetInactive"},
  8: {message:"DistributionInactive"},
  9: {message:"BuyerNotEligible"},
  10: {message:"InvalidAmount"},
  11: {message:"InvalidExpiry"},
  12: {message:"OrderNotFound"},
  13: {message:"InvalidOrderStatus"},
  14: {message:"PaymentAlreadyFunded"},
  15: {message:"AssetAlreadyFunded"},
  16: {message:"OrderExpired"},
  17: {message:"NotExpired"},
  18: {message:"NotCancellable"},
  19: {message:"NothingToCancel"}
}


export interface OrderRecord {
  asset: string;
  asset_amount: i128;
  asset_funded: boolean;
  buyer: string;
  created_at_ledger: u32;
  distributor: string;
  expires_at_ledger: u32;
  id: u64;
  payment_amount: i128;
  payment_asset: string;
  payment_funded: boolean;
  status: OrderStatus;
}

export type OrderStatus = {tag: "Created", values: void} | {tag: "Settled", values: void} | {tag: "Cancelled", values: void} | {tag: "Expired", values: void};












export type DataKey = {tag: "Admin", values: void} | {tag: "PendingAdmin", values: void} | {tag: "Initialized", values: void} | {tag: "Registry", values: void} | {tag: "Paused", values: void} | {tag: "NextOrderId", values: void} | {tag: "Order", values: readonly [u64]};

export interface Client {
  /**
   * Construct and simulate a admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  admin: (options?: MethodOptions) => Promise<AssembledTransaction<string>>

  /**
   * Construct and simulate a pause transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  pause: (options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a settle transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  settle: ({order_id}: {order_id: u64}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a unpause transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  unpause: (options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a registry transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  registry: (options?: MethodOptions) => Promise<AssembledTransaction<string>>

  /**
   * Construct and simulate a get_order transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  get_order: ({order_id}: {order_id: u64}, options?: MethodOptions) => Promise<AssembledTransaction<Option<OrderRecord>>>

  /**
   * Construct and simulate a is_paused transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  is_paused: (options?: MethodOptions) => Promise<AssembledTransaction<boolean>>

  /**
   * Construct and simulate a fund_asset transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  fund_asset: ({order_id}: {order_id: u64}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a initialize transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  initialize: ({admin, registry}: {admin: string, registry: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a accept_admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  accept_admin: (options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a cancel_order transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  cancel_order: ({order_id}: {order_id: u64}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a create_order transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  create_order: ({distributor, buyer, asset, payment_asset, asset_amount, payment_amount, expires_at_ledger}: {distributor: string, buyer: string, asset: string, payment_asset: string, asset_amount: i128, payment_amount: i128, expires_at_ledger: u32}, options?: MethodOptions) => Promise<AssembledTransaction<Result<u64>>>

  /**
   * Construct and simulate a expire_order transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  expire_order: ({order_id}: {order_id: u64}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a fund_payment transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  fund_payment: ({order_id}: {order_id: u64}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a order_exists transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  order_exists: ({order_id}: {order_id: u64}, options?: MethodOptions) => Promise<AssembledTransaction<boolean>>

  /**
   * Construct and simulate a next_order_id transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  next_order_id: (options?: MethodOptions) => Promise<AssembledTransaction<u64>>

  /**
   * Construct and simulate a pending_admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  pending_admin: (options?: MethodOptions) => Promise<AssembledTransaction<Option<string>>>

  /**
   * Construct and simulate a propose_admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  propose_admin: ({new_admin}: {new_admin: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a cancel_admin_proposal transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  cancel_admin_proposal: (options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

}
export class Client extends ContractClient {
  static async deploy<T = Client>(
    /** Options for initializing a Client as well as for calling a method, with extras specific to deploying. */
    options: MethodOptions &
      Omit<ContractClientOptions, "contractId"> & {
        /** The hash of the Wasm blob, which must already be installed on-chain. */
        wasmHash: Buffer | string;
        /** Salt used to generate the contract's ID. Passed through to {@link Operation.createCustomContract}. Default: random. */
        salt?: Buffer | Uint8Array;
        /** The format used to decode `wasmHash`, if it's provided as a string. */
        format?: "hex" | "base64";
      }
  ): Promise<AssembledTransaction<T>> {
    return ContractClient.deploy(null, options)
  }
  constructor(public readonly options: ContractClientOptions) {
    super(
      new ContractSpec([ "AAAAAQAAAAAAAAAAAAAAEUVsaWdpYmlsaXR5UmVjb3JkAAAAAAAABAAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAAAAAAVidXllcgAAAAAAABMAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAAAAAAAEnZhbGlkX3VudGlsX2xlZGdlcgAAAAAABA==",
        "AAAAAAAAAAAAAAAFYWRtaW4AAAAAAAAAAAAAAQAAABM=",
        "AAAAAAAAAAAAAAAFcGF1c2UAAAAAAAAAAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAGc2V0dGxlAAAAAAABAAAAAAAAAAhvcmRlcl9pZAAAAAYAAAABAAAD6QAAAAIAAAfQAAAADUNvbnRyYWN0RXJyb3IAAAA=",
        "AAAAAAAAAAAAAAAHdW5wYXVzZQAAAAAAAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAIcmVnaXN0cnkAAAAAAAAAAQAAABM=",
        "AAAAAAAAAAAAAAAJZ2V0X29yZGVyAAAAAAAAAQAAAAAAAAAIb3JkZXJfaWQAAAAGAAAAAQAAA+gAAAfQAAAAC09yZGVyUmVjb3JkAA==",
        "AAAAAAAAAAAAAAAJaXNfcGF1c2VkAAAAAAAAAAAAAAEAAAAB",
        "AAAAAAAAAAAAAAAKZnVuZF9hc3NldAAAAAAAAQAAAAAAAAAIb3JkZXJfaWQAAAAGAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAKaW5pdGlhbGl6ZQAAAAAAAgAAAAAAAAAFYWRtaW4AAAAAAAATAAAAAAAAAAhyZWdpc3RyeQAAABMAAAABAAAD6QAAAAIAAAfQAAAADUNvbnRyYWN0RXJyb3IAAAA=",
        "AAAAAAAAAAAAAAAMYWNjZXB0X2FkbWluAAAAAAAAAAEAAAPpAAAAAgAAB9AAAAANQ29udHJhY3RFcnJvcgAAAA==",
        "AAAAAAAAAAAAAAAMY2FuY2VsX29yZGVyAAAAAQAAAAAAAAAIb3JkZXJfaWQAAAAGAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAMY3JlYXRlX29yZGVyAAAABwAAAAAAAAALZGlzdHJpYnV0b3IAAAAAEwAAAAAAAAAFYnV5ZXIAAAAAAAATAAAAAAAAAAVhc3NldAAAAAAAABMAAAAAAAAADXBheW1lbnRfYXNzZXQAAAAAAAATAAAAAAAAAAxhc3NldF9hbW91bnQAAAALAAAAAAAAAA5wYXltZW50X2Ftb3VudAAAAAAACwAAAAAAAAARZXhwaXJlc19hdF9sZWRnZXIAAAAAAAAEAAAAAQAAA+kAAAAGAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAMZXhwaXJlX29yZGVyAAAAAQAAAAAAAAAIb3JkZXJfaWQAAAAGAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAMZnVuZF9wYXltZW50AAAAAQAAAAAAAAAIb3JkZXJfaWQAAAAGAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAMb3JkZXJfZXhpc3RzAAAAAQAAAAAAAAAIb3JkZXJfaWQAAAAGAAAAAQAAAAE=",
        "AAAAAAAAAAAAAAANbmV4dF9vcmRlcl9pZAAAAAAAAAAAAAABAAAABg==",
        "AAAAAAAAAAAAAAANcGVuZGluZ19hZG1pbgAAAAAAAAAAAAABAAAD6AAAABM=",
        "AAAAAAAAAAAAAAANcHJvcG9zZV9hZG1pbgAAAAAAAAEAAAAAAAAACW5ld19hZG1pbgAAAAAAABMAAAABAAAD6QAAAAIAAAfQAAAADUNvbnRyYWN0RXJyb3IAAAA=",
        "AAAAAAAAAAAAAAAVY2FuY2VsX2FkbWluX3Byb3Bvc2FsAAAAAAAAAAAAAAEAAAPpAAAAAgAAB9AAAAANQ29udHJhY3RFcnJvcgAAAA==",
        "AAAABAAAAAAAAAAAAAAADUNvbnRyYWN0RXJyb3IAAAAAAAATAAAAAAAAAAxVbmF1dGhvcml6ZWQAAAABAAAAAAAAAA5Ob3RJbml0aWFsaXplZAAAAAAAAgAAAAAAAAASQWxyZWFkeUluaXRpYWxpemVkAAAAAAADAAAAAAAAABZQZW5kaW5nQWRtaW5BbHJlYWR5U2V0AAAAAAAEAAAAAAAAAA5Ob1BlbmRpbmdBZG1pbgAAAAAABQAAAAAAAAAGUGF1c2VkAAAAAAAGAAAAAAAAAA1Bc3NldEluYWN0aXZlAAAAAAAABwAAAAAAAAAURGlzdHJpYnV0aW9uSW5hY3RpdmUAAAAIAAAAAAAAABBCdXllck5vdEVsaWdpYmxlAAAACQAAAAAAAAANSW52YWxpZEFtb3VudAAAAAAAAAoAAAAAAAAADUludmFsaWRFeHBpcnkAAAAAAAALAAAAAAAAAA1PcmRlck5vdEZvdW5kAAAAAAAADAAAAAAAAAASSW52YWxpZE9yZGVyU3RhdHVzAAAAAAANAAAAAAAAABRQYXltZW50QWxyZWFkeUZ1bmRlZAAAAA4AAAAAAAAAEkFzc2V0QWxyZWFkeUZ1bmRlZAAAAAAADwAAAAAAAAAMT3JkZXJFeHBpcmVkAAAAEAAAAAAAAAAKTm90RXhwaXJlZAAAAAAAEQAAAAAAAAAOTm90Q2FuY2VsbGFibGUAAAAAABIAAAAAAAAAD05vdGhpbmdUb0NhbmNlbAAAAAAT",
        "AAAAAQAAAAAAAAAAAAAAC09yZGVyUmVjb3JkAAAAAAwAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAAAAAAMYXNzZXRfYW1vdW50AAAACwAAAAAAAAAMYXNzZXRfZnVuZGVkAAAAAQAAAAAAAAAFYnV5ZXIAAAAAAAATAAAAAAAAABFjcmVhdGVkX2F0X2xlZGdlcgAAAAAAAAQAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAAAAAAAEWV4cGlyZXNfYXRfbGVkZ2VyAAAAAAAABAAAAAAAAAACaWQAAAAAAAYAAAAAAAAADnBheW1lbnRfYW1vdW50AAAAAAALAAAAAAAAAA1wYXltZW50X2Fzc2V0AAAAAAAAEwAAAAAAAAAOcGF5bWVudF9mdW5kZWQAAAAAAAEAAAAAAAAABnN0YXR1cwAAAAAH0AAAAAtPcmRlclN0YXR1cwA=",
        "AAAAAgAAAAAAAAAAAAAAC09yZGVyU3RhdHVzAAAAAAQAAAAAAAAAAAAAAAdDcmVhdGVkAAAAAAAAAAAAAAAAB1NldHRsZWQAAAAAAAAAAAAAAAAJQ2FuY2VsbGVkAAAAAAAAAAAAAAAAAAAHRXhwaXJlZAA=",
        "AAAABQAAAAAAAAAAAAAAC0Fzc2V0RnVuZGVkAAAAAAEAAAAMYXNzZXRfZnVuZGVkAAAAAwAAAAAAAAAIb3JkZXJfaWQAAAAGAAAAAQAAAAAAAAALZGlzdHJpYnV0b3IAAAAAEwAAAAEAAAAAAAAADGFzc2V0X2Ftb3VudAAAAAsAAAAAAAAAAg==",
        "AAAABQAAAAAAAAAAAAAADE9yZGVyQ3JlYXRlZAAAAAEAAAANb3JkZXJfY3JlYXRlZAAAAAAAAAgAAAAAAAAACG9yZGVyX2lkAAAABgAAAAEAAAAAAAAABWJ1eWVyAAAAAAAAEwAAAAEAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAABAAAAAAAAAAVhc3NldAAAAAAAABMAAAAAAAAAAAAAAA1wYXltZW50X2Fzc2V0AAAAAAAAEwAAAAAAAAAAAAAADGFzc2V0X2Ftb3VudAAAAAsAAAAAAAAAAAAAAA5wYXltZW50X2Ftb3VudAAAAAAACwAAAAAAAAAAAAAAEWV4cGlyZXNfYXRfbGVkZ2VyAAAAAAAABAAAAAAAAAAC",
        "AAAABQAAAAAAAAAAAAAADE9yZGVyRXhwaXJlZAAAAAEAAAANb3JkZXJfZXhwaXJlZAAAAAAAAAMAAAAAAAAACG9yZGVyX2lkAAAABgAAAAEAAAAAAAAAEHBheW1lbnRfcmVmdW5kZWQAAAABAAAAAAAAAAAAAAAOYXNzZXRfcmVmdW5kZWQAAAAAAAEAAAAAAAAAAg==",
        "AAAABQAAAAAAAAAAAAAADE9yZGVyU2V0dGxlZAAAAAEAAAANb3JkZXJfc2V0dGxlZAAAAAAAAAUAAAAAAAAACG9yZGVyX2lkAAAABgAAAAEAAAAAAAAABWJ1eWVyAAAAAAAAEwAAAAEAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAABAAAAAAAAAAxhc3NldF9hbW91bnQAAAALAAAAAAAAAAAAAAAOcGF5bWVudF9hbW91bnQAAAAAAAsAAAAAAAAAAg==",
        "AAAABQAAAAAAAAAAAAAADUdhdGV3YXlQYXVzZWQAAAAAAAABAAAADmdhdGV3YXlfcGF1c2VkAAAAAAAAAAAAAg==",
        "AAAABQAAAAAAAAAAAAAADVBheW1lbnRGdW5kZWQAAAAAAAABAAAADnBheW1lbnRfZnVuZGVkAAAAAAADAAAAAAAAAAhvcmRlcl9pZAAAAAYAAAABAAAAAAAAAAVidXllcgAAAAAAABMAAAABAAAAAAAAAA5wYXltZW50X2Ftb3VudAAAAAAACwAAAAAAAAAC",
        "AAAABQAAAAAAAAAAAAAADk9yZGVyQ2FuY2VsbGVkAAAAAAABAAAAD29yZGVyX2NhbmNlbGxlZAAAAAACAAAAAAAAAAhvcmRlcl9pZAAAAAYAAAABAAAAAAAAAAxjYW5jZWxsZWRfYnkAAAATAAAAAQAAAAI=",
        "AAAABQAAAAAAAAAAAAAAD0dhdGV3YXlVbnBhdXNlZAAAAAABAAAAEGdhdGV3YXlfdW5wYXVzZWQAAAAAAAAAAg==",
        "AAAABQAAAAAAAAAAAAAAEEFkbWluVHJhbnNmZXJyZWQAAAABAAAAEWFkbWluX3RyYW5zZmVycmVkAAAAAAAAAgAAAAAAAAAOcHJldmlvdXNfYWRtaW4AAAAAABMAAAABAAAAAAAAAAluZXdfYWRtaW4AAAAAAAATAAAAAAAAAAI=",
        "AAAABQAAAAAAAAAAAAAAFUFkbWluVHJhbnNmZXJQcm9wb3NlZAAAAAAAAAEAAAAXYWRtaW5fdHJhbnNmZXJfcHJvcG9zZWQAAAAAAQAAAAAAAAAJbmV3X2FkbWluAAAAAAAAEwAAAAEAAAAC",
        "AAAABQAAAAAAAAAAAAAAFkFkbWluVHJhbnNmZXJDYW5jZWxsZWQAAAAAAAEAAAAYYWRtaW5fdHJhbnNmZXJfY2FuY2VsbGVkAAAAAQAAAAAAAAANcGVuZGluZ19hZG1pbgAAAAAAABMAAAABAAAAAg==",
        "AAAAAgAAAAAAAAAAAAAAB0RhdGFLZXkAAAAABwAAAAAAAAAAAAAABUFkbWluAAAAAAAAAAAAAAAAAAAMUGVuZGluZ0FkbWluAAAAAAAAAAAAAAALSW5pdGlhbGl6ZWQAAAAAAAAAAAAAAAAIUmVnaXN0cnkAAAAAAAAAAAAAAAZQYXVzZWQAAAAAAAAAAAAAAAAAC05leHRPcmRlcklkAAAAAAEAAAAAAAAABU9yZGVyAAAAAAAAAQAAAAY=" ]),
      options
    )
  }
  public readonly fromJSON = {
    admin: this.txFromJSON<string>,
        pause: this.txFromJSON<Result<void>>,
        settle: this.txFromJSON<Result<void>>,
        unpause: this.txFromJSON<Result<void>>,
        registry: this.txFromJSON<string>,
        get_order: this.txFromJSON<Option<OrderRecord>>,
        is_paused: this.txFromJSON<boolean>,
        fund_asset: this.txFromJSON<Result<void>>,
        initialize: this.txFromJSON<Result<void>>,
        accept_admin: this.txFromJSON<Result<void>>,
        cancel_order: this.txFromJSON<Result<void>>,
        create_order: this.txFromJSON<Result<u64>>,
        expire_order: this.txFromJSON<Result<void>>,
        fund_payment: this.txFromJSON<Result<void>>,
        order_exists: this.txFromJSON<boolean>,
        next_order_id: this.txFromJSON<u64>,
        pending_admin: this.txFromJSON<Option<string>>,
        propose_admin: this.txFromJSON<Result<void>>,
        cancel_admin_proposal: this.txFromJSON<Result<void>>
  }
}