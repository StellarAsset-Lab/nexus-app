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




export const ContractError = {
  1: {message:"Unauthorized"},
  2: {message:"NotInitialized"},
  3: {message:"AlreadyInitialized"},
  4: {message:"PendingAdminAlreadySet"},
  5: {message:"NoPendingAdmin"},
  6: {message:"AssetNotFound"},
  7: {message:"AssetAlreadyActive"},
  8: {message:"IssuerMismatch"},
  9: {message:"DistributionNotFound"},
  10: {message:"DistributionAlreadyActive"},
  11: {message:"EligibilityAuthorityMismatch"},
  12: {message:"EligibilityNotFound"},
  13: {message:"InvalidEligibilityExpiry"},
  14: {message:"AssetInactive"},
  15: {message:"DistributionInactive"}
}


export interface AssetRecord {
  active: boolean;
  asset: string;
  issuer: string;
}


export interface EligibilityRecord {
  asset: string;
  buyer: string;
  distributor: string;
  valid_until_ledger: u32;
}


export interface DistributionConfig {
  active: boolean;
  asset: string;
  distributor: string;
  eligibility_authority: string;
}










export type DataKey = {tag: "Asset", values: readonly [string]} | {tag: "Distribution", values: readonly [string, string]} | {tag: "Eligibility", values: readonly [string, string, string]} | {tag: "Admin", values: void} | {tag: "PendingAdmin", values: void} | {tag: "Initialized", values: void};

export interface Client {
  /**
   * Construct and simulate a admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  admin: (options?: MethodOptions) => Promise<AssembledTransaction<string>>

  /**
   * Construct and simulate a get_asset transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  get_asset: ({asset}: {asset: string}, options?: MethodOptions) => Promise<AssembledTransaction<Option<AssetRecord>>>

  /**
   * Construct and simulate a initialize transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  initialize: ({admin}: {admin: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a is_eligible transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  is_eligible: ({asset, distributor, buyer}: {asset: string, distributor: string, buyer: string}, options?: MethodOptions) => Promise<AssembledTransaction<boolean>>

  /**
   * Construct and simulate a accept_admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  accept_admin: (options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a pending_admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  pending_admin: (options?: MethodOptions) => Promise<AssembledTransaction<Option<string>>>

  /**
   * Construct and simulate a propose_admin transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  propose_admin: ({new_admin}: {new_admin: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a register_asset transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  register_asset: ({asset, issuer}: {asset: string, issuer: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a get_eligibility transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  get_eligibility: ({asset, distributor, buyer}: {asset: string, distributor: string, buyer: string}, options?: MethodOptions) => Promise<AssembledTransaction<Option<EligibilityRecord>>>

  /**
   * Construct and simulate a is_asset_active transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  is_asset_active: ({asset}: {asset: string}, options?: MethodOptions) => Promise<AssembledTransaction<boolean>>

  /**
   * Construct and simulate a set_eligibility transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  set_eligibility: ({asset, distributor, buyer, valid_until_ledger}: {asset: string, distributor: string, buyer: string, valid_until_ledger: u32}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a deactivate_asset transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  deactivate_asset: ({asset}: {asset: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a get_distribution transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  get_distribution: ({asset, distributor}: {asset: string, distributor: string}, options?: MethodOptions) => Promise<AssembledTransaction<Option<DistributionConfig>>>

  /**
   * Construct and simulate a revoke_eligibility transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  revoke_eligibility: ({asset, distributor, buyer}: {asset: string, distributor: string, buyer: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a revoke_distribution transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  revoke_distribution: ({asset, distributor}: {asset: string, distributor: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a cancel_admin_proposal transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  cancel_admin_proposal: (options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a register_distribution transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  register_distribution: ({asset, distributor, eligibility_authority}: {asset: string, distributor: string, eligibility_authority: string}, options?: MethodOptions) => Promise<AssembledTransaction<Result<void>>>

  /**
   * Construct and simulate a is_distribution_active transaction. Returns an `AssembledTransaction` object which will have a `result` field containing the result of the simulation. If this transaction changes contract state, you will need to call `signAndSend()` on the returned object.
   */
  is_distribution_active: ({asset, distributor}: {asset: string, distributor: string}, options?: MethodOptions) => Promise<AssembledTransaction<boolean>>

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
      new ContractSpec([ "AAAAAAAAAAAAAAAFYWRtaW4AAAAAAAAAAAAAAQAAABM=",
        "AAAAAAAAAAAAAAAJZ2V0X2Fzc2V0AAAAAAAAAQAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAQAAA+gAAAfQAAAAC0Fzc2V0UmVjb3JkAA==",
        "AAAAAAAAAAAAAAAKaW5pdGlhbGl6ZQAAAAAAAQAAAAAAAAAFYWRtaW4AAAAAAAATAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAALaXNfZWxpZ2libGUAAAAAAwAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAAAAAAtkaXN0cmlidXRvcgAAAAATAAAAAAAAAAVidXllcgAAAAAAABMAAAABAAAAAQ==",
        "AAAAAAAAAAAAAAAMYWNjZXB0X2FkbWluAAAAAAAAAAEAAAPpAAAAAgAAB9AAAAANQ29udHJhY3RFcnJvcgAAAA==",
        "AAAAAAAAAAAAAAANcGVuZGluZ19hZG1pbgAAAAAAAAAAAAABAAAD6AAAABM=",
        "AAAAAAAAAAAAAAANcHJvcG9zZV9hZG1pbgAAAAAAAAEAAAAAAAAACW5ld19hZG1pbgAAAAAAABMAAAABAAAD6QAAAAIAAAfQAAAADUNvbnRyYWN0RXJyb3IAAAA=",
        "AAAAAAAAAAAAAAAOcmVnaXN0ZXJfYXNzZXQAAAAAAAIAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAAAAAAGaXNzdWVyAAAAAAATAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAPZ2V0X2VsaWdpYmlsaXR5AAAAAAMAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAAAAAALZGlzdHJpYnV0b3IAAAAAEwAAAAAAAAAFYnV5ZXIAAAAAAAATAAAAAQAAA+gAAAfQAAAAEUVsaWdpYmlsaXR5UmVjb3JkAAAA",
        "AAAAAAAAAAAAAAAPaXNfYXNzZXRfYWN0aXZlAAAAAAEAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAEAAAAB",
        "AAAAAAAAAAAAAAAPc2V0X2VsaWdpYmlsaXR5AAAAAAQAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAAAAAALZGlzdHJpYnV0b3IAAAAAEwAAAAAAAAAFYnV5ZXIAAAAAAAATAAAAAAAAABJ2YWxpZF91bnRpbF9sZWRnZXIAAAAAAAQAAAABAAAD6QAAAAIAAAfQAAAADUNvbnRyYWN0RXJyb3IAAAA=",
        "AAAAAAAAAAAAAAAQZGVhY3RpdmF0ZV9hc3NldAAAAAEAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAEAAAPpAAAAAgAAB9AAAAANQ29udHJhY3RFcnJvcgAAAA==",
        "AAAAAAAAAAAAAAAQZ2V0X2Rpc3RyaWJ1dGlvbgAAAAIAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAAAAAALZGlzdHJpYnV0b3IAAAAAEwAAAAEAAAPoAAAH0AAAABJEaXN0cmlidXRpb25Db25maWcAAA==",
        "AAAAAAAAAAAAAAAScmV2b2tlX2VsaWdpYmlsaXR5AAAAAAADAAAAAAAAAAVhc3NldAAAAAAAABMAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAAAAAAABWJ1eWVyAAAAAAAAEwAAAAEAAAPpAAAAAgAAB9AAAAANQ29udHJhY3RFcnJvcgAAAA==",
        "AAAAAAAAAAAAAAATcmV2b2tlX2Rpc3RyaWJ1dGlvbgAAAAACAAAAAAAAAAVhc3NldAAAAAAAABMAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAABAAAD6QAAAAIAAAfQAAAADUNvbnRyYWN0RXJyb3IAAAA=",
        "AAAAAAAAAAAAAAAVY2FuY2VsX2FkbWluX3Byb3Bvc2FsAAAAAAAAAAAAAAEAAAPpAAAAAgAAB9AAAAANQ29udHJhY3RFcnJvcgAAAA==",
        "AAAAAAAAAAAAAAAVcmVnaXN0ZXJfZGlzdHJpYnV0aW9uAAAAAAAAAwAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAAAAAAtkaXN0cmlidXRvcgAAAAATAAAAAAAAABVlbGlnaWJpbGl0eV9hdXRob3JpdHkAAAAAAAATAAAAAQAAA+kAAAACAAAH0AAAAA1Db250cmFjdEVycm9yAAAA",
        "AAAAAAAAAAAAAAAWaXNfZGlzdHJpYnV0aW9uX2FjdGl2ZQAAAAAAAgAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAAAAAAtkaXN0cmlidXRvcgAAAAATAAAAAQAAAAE=",
        "AAAABAAAAAAAAAAAAAAADUNvbnRyYWN0RXJyb3IAAAAAAAAPAAAAAAAAAAxVbmF1dGhvcml6ZWQAAAABAAAAAAAAAA5Ob3RJbml0aWFsaXplZAAAAAAAAgAAAAAAAAASQWxyZWFkeUluaXRpYWxpemVkAAAAAAADAAAAAAAAABZQZW5kaW5nQWRtaW5BbHJlYWR5U2V0AAAAAAAEAAAAAAAAAA5Ob1BlbmRpbmdBZG1pbgAAAAAABQAAAAAAAAANQXNzZXROb3RGb3VuZAAAAAAAAAYAAAAAAAAAEkFzc2V0QWxyZWFkeUFjdGl2ZQAAAAAABwAAAAAAAAAOSXNzdWVyTWlzbWF0Y2gAAAAAAAgAAAAAAAAAFERpc3RyaWJ1dGlvbk5vdEZvdW5kAAAACQAAAAAAAAAZRGlzdHJpYnV0aW9uQWxyZWFkeUFjdGl2ZQAAAAAAAAoAAAAAAAAAHEVsaWdpYmlsaXR5QXV0aG9yaXR5TWlzbWF0Y2gAAAALAAAAAAAAABNFbGlnaWJpbGl0eU5vdEZvdW5kAAAAAAwAAAAAAAAAGEludmFsaWRFbGlnaWJpbGl0eUV4cGlyeQAAAA0AAAAAAAAADUFzc2V0SW5hY3RpdmUAAAAAAAAOAAAAAAAAABREaXN0cmlidXRpb25JbmFjdGl2ZQAAAA8=",
        "AAAAAQAAAAAAAAAAAAAAC0Fzc2V0UmVjb3JkAAAAAAMAAAAAAAAABmFjdGl2ZQAAAAAAAQAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAAAAAAZpc3N1ZXIAAAAAABM=",
        "AAAAAQAAAAAAAAAAAAAAEUVsaWdpYmlsaXR5UmVjb3JkAAAAAAAABAAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAAAAAAVidXllcgAAAAAAABMAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAAAAAAAEnZhbGlkX3VudGlsX2xlZGdlcgAAAAAABA==",
        "AAAAAQAAAAAAAAAAAAAAEkRpc3RyaWJ1dGlvbkNvbmZpZwAAAAAABAAAAAAAAAAGYWN0aXZlAAAAAAABAAAAAAAAAAVhc3NldAAAAAAAABMAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAAAAAAAFWVsaWdpYmlsaXR5X2F1dGhvcml0eQAAAAAAABM=",
        "AAAABQAAAAAAAAAAAAAADkVsaWdpYmlsaXR5U2V0AAAAAAABAAAAD2VsaWdpYmlsaXR5X3NldAAAAAAEAAAAAAAAAAVhc3NldAAAAAAAABMAAAABAAAAAAAAAAtkaXN0cmlidXRvcgAAAAATAAAAAQAAAAAAAAAFYnV5ZXIAAAAAAAATAAAAAQAAAAAAAAASdmFsaWRfdW50aWxfbGVkZ2VyAAAAAAAEAAAAAAAAAAI=",
        "AAAABQAAAAAAAAAAAAAAD0Fzc2V0UmVnaXN0ZXJlZAAAAAABAAAAEGFzc2V0X3JlZ2lzdGVyZWQAAAACAAAAAAAAAAVhc3NldAAAAAAAABMAAAABAAAAAAAAAAZpc3N1ZXIAAAAAABMAAAAAAAAAAg==",
        "AAAABQAAAAAAAAAAAAAAEEFkbWluVHJhbnNmZXJyZWQAAAABAAAAEWFkbWluX3RyYW5zZmVycmVkAAAAAAAAAgAAAAAAAAAOcHJldmlvdXNfYWRtaW4AAAAAABMAAAABAAAAAAAAAAluZXdfYWRtaW4AAAAAAAATAAAAAAAAAAI=",
        "AAAABQAAAAAAAAAAAAAAEEFzc2V0RGVhY3RpdmF0ZWQAAAABAAAAEWFzc2V0X2RlYWN0aXZhdGVkAAAAAAAAAQAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAQAAAAI=",
        "AAAABQAAAAAAAAAAAAAAEkVsaWdpYmlsaXR5UmV2b2tlZAAAAAAAAQAAABNlbGlnaWJpbGl0eV9yZXZva2VkAAAAAAMAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAEAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAABAAAAAAAAAAVidXllcgAAAAAAABMAAAABAAAAAg==",
        "AAAABQAAAAAAAAAAAAAAE0Rpc3RyaWJ1dGlvblJldm9rZWQAAAAAAQAAABRkaXN0cmlidXRpb25fcmV2b2tlZAAAAAIAAAAAAAAABWFzc2V0AAAAAAAAEwAAAAEAAAAAAAAAC2Rpc3RyaWJ1dG9yAAAAABMAAAABAAAAAg==",
        "AAAABQAAAAAAAAAAAAAAFUFkbWluVHJhbnNmZXJQcm9wb3NlZAAAAAAAAAEAAAAXYWRtaW5fdHJhbnNmZXJfcHJvcG9zZWQAAAAAAQAAAAAAAAAJbmV3X2FkbWluAAAAAAAAEwAAAAEAAAAC",
        "AAAABQAAAAAAAAAAAAAAFkFkbWluVHJhbnNmZXJDYW5jZWxsZWQAAAAAAAEAAAAYYWRtaW5fdHJhbnNmZXJfY2FuY2VsbGVkAAAAAQAAAAAAAAANcGVuZGluZ19hZG1pbgAAAAAAABMAAAABAAAAAg==",
        "AAAABQAAAAAAAAAAAAAAFkRpc3RyaWJ1dGlvblJlZ2lzdGVyZWQAAAAAAAEAAAAXZGlzdHJpYnV0aW9uX3JlZ2lzdGVyZWQAAAAAAwAAAAAAAAAFYXNzZXQAAAAAAAATAAAAAQAAAAAAAAALZGlzdHJpYnV0b3IAAAAAEwAAAAEAAAAAAAAAFWVsaWdpYmlsaXR5X2F1dGhvcml0eQAAAAAAABMAAAAAAAAAAg==",
        "AAAAAgAAAAAAAAAAAAAAB0RhdGFLZXkAAAAABgAAAAEAAAAAAAAABUFzc2V0AAAAAAAAAQAAABMAAAABAAAAAAAAAAxEaXN0cmlidXRpb24AAAACAAAAEwAAABMAAAABAAAAAAAAAAtFbGlnaWJpbGl0eQAAAAADAAAAEwAAABMAAAATAAAAAAAAAAAAAAAFQWRtaW4AAAAAAAAAAAAAAAAAAAxQZW5kaW5nQWRtaW4AAAAAAAAAAAAAAAtJbml0aWFsaXplZAA=" ]),
      options
    )
  }
  public readonly fromJSON = {
    admin: this.txFromJSON<string>,
        get_asset: this.txFromJSON<Option<AssetRecord>>,
        initialize: this.txFromJSON<Result<void>>,
        is_eligible: this.txFromJSON<boolean>,
        accept_admin: this.txFromJSON<Result<void>>,
        pending_admin: this.txFromJSON<Option<string>>,
        propose_admin: this.txFromJSON<Result<void>>,
        register_asset: this.txFromJSON<Result<void>>,
        get_eligibility: this.txFromJSON<Option<EligibilityRecord>>,
        is_asset_active: this.txFromJSON<boolean>,
        set_eligibility: this.txFromJSON<Result<void>>,
        deactivate_asset: this.txFromJSON<Result<void>>,
        get_distribution: this.txFromJSON<Option<DistributionConfig>>,
        revoke_eligibility: this.txFromJSON<Result<void>>,
        revoke_distribution: this.txFromJSON<Result<void>>,
        cancel_admin_proposal: this.txFromJSON<Result<void>>,
        register_distribution: this.txFromJSON<Result<void>>,
        is_distribution_active: this.txFromJSON<boolean>
  }
}