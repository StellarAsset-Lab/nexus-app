package indexing

import (
	"crypto/rand"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// The helpers below build real xdr.ScVal values the same way the actual
// Soroban SDK would encode them, so these tests exercise the genuine wire
// format rather than a decoder-authored fixture shape.

func testAddressScVal(t *testing.T) (address string, scVal xdr.ScVal) {
	t.Helper()
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		t.Fatalf("random bytes: %v", err)
	}
	contractID := xdr.ContractId(xdr.Hash(raw))
	scAddr, err := xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, contractID)
	if err != nil {
		t.Fatalf("new sc address: %v", err)
	}
	v, err := xdr.NewScVal(xdr.ScValTypeScvAddress, scAddr)
	if err != nil {
		t.Fatalf("new sc val: %v", err)
	}
	addrStr, err := scAddr.String()
	if err != nil {
		t.Fatalf("address string: %v", err)
	}
	return addrStr, v
}

func testSymbolScVal(t *testing.T, s string) xdr.ScVal {
	t.Helper()
	v, err := xdr.NewScVal(xdr.ScValTypeScvSymbol, xdr.ScSymbol(s))
	if err != nil {
		t.Fatalf("new symbol scval: %v", err)
	}
	return v
}

func testU32ScVal(t *testing.T, n uint32) xdr.ScVal {
	t.Helper()
	v, err := xdr.NewScVal(xdr.ScValTypeScvU32, xdr.Uint32(n))
	if err != nil {
		t.Fatalf("new u32 scval: %v", err)
	}
	return v
}

// testMapScVal builds a real Map<Symbol, Val> ScVal, matching how Soroban's
// #[contractevent] macro (data_format defaults to "map" regardless of field
// count — confirmed against the real compiled ABI) actually encodes event
// data fields.
func testMapScVal(t *testing.T, fields map[string]xdr.ScVal) xdr.ScVal {
	t.Helper()
	entries := make(xdr.ScMap, 0, len(fields))
	for k, v := range fields {
		key, err := xdr.NewScVal(xdr.ScValTypeScvSymbol, xdr.ScSymbol(k))
		if err != nil {
			t.Fatalf("new symbol key: %v", err)
		}
		entries = append(entries, xdr.ScMapEntry{Key: key, Val: v})
	}
	v, err := xdr.NewScVal(xdr.ScValTypeScvMap, &entries)
	if err != nil {
		t.Fatalf("new map scval: %v", err)
	}
	return v
}

func TestDecodeRegistryEvent_AssetRegistered(t *testing.T) {
	asset, assetVal := testAddressScVal(t)
	issuer, issuerVal := testAddressScVal(t)

	topics := []xdr.ScVal{testSymbolScVal(t, EventAssetRegistered), assetVal}
	payload, err := DecodeRegistryEvent(EventAssetRegistered, topics, testMapScVal(t, map[string]xdr.ScVal{"issuer": issuerVal}))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["asset"] != asset {
		t.Fatalf("expected asset %s, got %v", asset, payload["asset"])
	}
	if payload["issuer"] != issuer {
		t.Fatalf("expected issuer %s, got %v", issuer, payload["issuer"])
	}
}

func TestDecodeRegistryEvent_AssetDeactivated(t *testing.T) {
	asset, assetVal := testAddressScVal(t)
	topics := []xdr.ScVal{testSymbolScVal(t, EventAssetDeactivated), assetVal}

	payload, err := DecodeRegistryEvent(EventAssetDeactivated, topics, xdr.ScVal{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["asset"] != asset {
		t.Fatalf("expected asset %s, got %v", asset, payload["asset"])
	}
}

func TestDecodeRegistryEvent_EligibilitySet(t *testing.T) {
	asset, assetVal := testAddressScVal(t)
	distributor, distributorVal := testAddressScVal(t)
	buyer, buyerVal := testAddressScVal(t)
	validUntil := uint32(123456)

	topics := []xdr.ScVal{testSymbolScVal(t, EventEligibilitySet), assetVal, distributorVal, buyerVal}
	payload, err := DecodeRegistryEvent(EventEligibilitySet, topics, testMapScVal(t, map[string]xdr.ScVal{"valid_until_ledger": testU32ScVal(t, validUntil)}))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["asset"] != asset || payload["distributor"] != distributor || payload["buyer"] != buyer {
		t.Fatalf("unexpected addresses in payload: %+v", payload)
	}
	if payload["valid_until_ledger"] != validUntil {
		t.Fatalf("expected valid_until_ledger=%d, got %v", validUntil, payload["valid_until_ledger"])
	}
}

func TestDecodeRegistryEvent_WrongTopicCountFails(t *testing.T) {
	_, assetVal := testAddressScVal(t)
	// AssetRegistered expects exactly 1 topic field; give it 2.
	_, extraVal := testAddressScVal(t)
	topics := []xdr.ScVal{testSymbolScVal(t, EventAssetRegistered), assetVal, extraVal}

	_, err := DecodeRegistryEvent(EventAssetRegistered, topics, xdr.ScVal{})
	if err == nil {
		t.Fatal("expected a decode error for the wrong topic count, got nil")
	}
}

func TestDecodeRegistryEvent_NonMapValueFails(t *testing.T) {
	_, assetVal := testAddressScVal(t)
	topics := []xdr.ScVal{testSymbolScVal(t, EventAssetRegistered), assetVal}

	// Every event's data is a Map (confirmed against the real ABI); a bare
	// scalar value must be rejected, not silently misread as a field.
	_, err := DecodeRegistryEvent(EventAssetRegistered, topics, testSymbolScVal(t, "not-a-map"))
	if err == nil {
		t.Fatal("expected a decode error for a non-Map value, got nil")
	}
}

func TestDecodeRegistryEvent_WrongFieldTypeWithinMapFails(t *testing.T) {
	_, assetVal := testAddressScVal(t)
	topics := []xdr.ScVal{testSymbolScVal(t, EventAssetRegistered), assetVal}

	// issuer should be an Address; give it a Symbol instead, inside an
	// otherwise well-formed data Map.
	badValue := testMapScVal(t, map[string]xdr.ScVal{"issuer": testSymbolScVal(t, "not-an-address")})
	_, err := DecodeRegistryEvent(EventAssetRegistered, topics, badValue)
	if err == nil {
		t.Fatal("expected a decode error for the wrong field type, got nil")
	}
}

func TestDecodeRegistryEvent_UnknownEventFails(t *testing.T) {
	_, assetVal := testAddressScVal(t)
	topics := []xdr.ScVal{testSymbolScVal(t, "totally_unknown_event"), assetVal}

	_, err := DecodeRegistryEvent("totally_unknown_event", topics, xdr.ScVal{})
	if err == nil {
		t.Fatal("expected a decode error for an unknown event name, got nil")
	}
}
