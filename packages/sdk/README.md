# @nexus/sdk

The human-facing TypeScript integration package for Nexus.

Wraps repetitive RPC, contract binding, wallet adapter, transaction lifecycle, and network configuration details without hiding important errors. Usable by any TypeScript application that wants to integrate Nexus.

Implementation lands incrementally across the initial commit sequence: network configuration, Registry reads, Order reads, order lifecycle writes, and typed error classification.
