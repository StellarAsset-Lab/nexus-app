package indexing

import (
	"math/big"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

func testU64ScVal(t *testing.T, n uint64) xdr.ScVal {
	t.Helper()
	v, err := xdr.NewScVal(xdr.ScValTypeScvU64, xdr.Uint64(n))
	if err != nil {
		t.Fatalf("new u64 scval: %v", err)
	}
	return v
}

func testBoolScVal(t *testing.T, b bool) xdr.ScVal {
	t.Helper()
	v, err := xdr.NewScVal(xdr.ScValTypeScvBool, b)
	if err != nil {
		t.Fatalf("new bool scval: %v", err)
	}
	return v
}

// testAmountScVal builds a real I128 ScVal for an exact big.Int amount,
// matching how the SDK actually splits i128 into Hi/Lo halves.
func testAmountScVal(t *testing.T, amount *big.Int) xdr.ScVal {
	t.Helper()
	mask64 := new(big.Int).SetUint64(^uint64(0))
	lo := new(big.Int).And(amount, mask64)
	hi := new(big.Int).Rsh(amount, 64)

	parts := xdr.Int128Parts{
		Hi: xdr.Int64(hi.Int64()),
		Lo: xdr.Uint64(lo.Uint64()),
	}
	v, err := xdr.NewScVal(xdr.ScValTypeScvI128, parts)
	if err != nil {
		t.Fatalf("new i128 scval: %v", err)
	}
	return v
}

func TestDecodeOrderEvent_OrderCreated_MultiFieldMap(t *testing.T) {
	// This is the exact case the earlier single-value shortcut bug would
	// have broken: 5 data fields, real Map encoding, real exact i128
	// amounts far beyond float64's safe integer range.
	orderID := uint64(42)
	buyer, buyerVal := testAddressScVal(t)
	distributor, distributorVal := testAddressScVal(t)
	asset, assetVal := testAddressScVal(t)
	paymentAsset, paymentAssetVal := testAddressScVal(t)
	assetAmount, _ := new(big.Int).SetString("170141183460469231731687303715884105727", 10) // i128 max
	paymentAmount := big.NewInt(9_000_000_000_000)
	expiresAtLedger := uint32(99999)

	topics := []xdr.ScVal{
		testSymbolScVal(t, EventOrderCreated),
		testU64ScVal(t, orderID), buyerVal, distributorVal,
	}
	value := testMapScVal(t, map[string]xdr.ScVal{
		"asset":             assetVal,
		"payment_asset":     paymentAssetVal,
		"asset_amount":      testAmountScVal(t, assetAmount),
		"payment_amount":    testAmountScVal(t, paymentAmount),
		"expires_at_ledger": testU32ScVal(t, expiresAtLedger),
	})

	payload, err := DecodeOrderEvent(EventOrderCreated, topics, value)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if payload["order_id"] != orderID {
		t.Fatalf("expected order_id=%d, got %v", orderID, payload["order_id"])
	}
	if payload["buyer"] != buyer || payload["distributor"] != distributor {
		t.Fatalf("unexpected addresses: %+v", payload)
	}
	if payload["asset"] != asset || payload["payment_asset"] != paymentAsset {
		t.Fatalf("unexpected asset addresses: %+v", payload)
	}
	gotAssetAmount, ok := payload["asset_amount"].(model.Amount)
	if !ok || gotAssetAmount.String() != assetAmount.String() {
		t.Fatalf("expected exact asset_amount=%s, got %v", assetAmount, payload["asset_amount"])
	}
	gotPaymentAmount, ok := payload["payment_amount"].(model.Amount)
	if !ok || gotPaymentAmount.String() != paymentAmount.String() {
		t.Fatalf("expected exact payment_amount=%s, got %v", paymentAmount, payload["payment_amount"])
	}
	if payload["expires_at_ledger"] != expiresAtLedger {
		t.Fatalf("expected expires_at_ledger=%d, got %v", expiresAtLedger, payload["expires_at_ledger"])
	}
}

func TestDecodeOrderEvent_OrderSettled(t *testing.T) {
	orderID := uint64(7)
	buyer, buyerVal := testAddressScVal(t)
	distributor, distributorVal := testAddressScVal(t)
	assetAmount := big.NewInt(123456789)
	paymentAmount := big.NewInt(987654321)

	topics := []xdr.ScVal{testSymbolScVal(t, EventOrderSettled), testU64ScVal(t, orderID), buyerVal, distributorVal}
	value := testMapScVal(t, map[string]xdr.ScVal{
		"asset_amount":   testAmountScVal(t, assetAmount),
		"payment_amount": testAmountScVal(t, paymentAmount),
	})

	payload, err := DecodeOrderEvent(EventOrderSettled, topics, value)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["order_id"] != orderID || payload["buyer"] != buyer || payload["distributor"] != distributor {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestDecodeOrderEvent_OrderExpired_BoolFields(t *testing.T) {
	orderID := uint64(3)
	topics := []xdr.ScVal{testSymbolScVal(t, EventOrderExpired), testU64ScVal(t, orderID)}
	value := testMapScVal(t, map[string]xdr.ScVal{
		"payment_refunded": testBoolScVal(t, true),
		"asset_refunded":   testBoolScVal(t, false),
	})

	payload, err := DecodeOrderEvent(EventOrderExpired, topics, value)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["payment_refunded"] != true || payload["asset_refunded"] != false {
		t.Fatalf("unexpected bool fields: %+v", payload)
	}
}

func TestDecodeOrderEvent_GatewayPaused_NoFields(t *testing.T) {
	topics := []xdr.ScVal{testSymbolScVal(t, EventGatewayPaused)}
	payload, err := DecodeOrderEvent(EventGatewayPaused, topics, xdr.ScVal{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload) != 0 {
		t.Fatalf("expected empty payload, got %+v", payload)
	}
}

func TestDecodeOrderEvent_PaymentFunded(t *testing.T) {
	orderID := uint64(15)
	buyer, buyerVal := testAddressScVal(t)
	paymentAmount := big.NewInt(5000)

	topics := []xdr.ScVal{testSymbolScVal(t, EventPaymentFunded), testU64ScVal(t, orderID), buyerVal}
	value := testMapScVal(t, map[string]xdr.ScVal{"payment_amount": testAmountScVal(t, paymentAmount)})

	payload, err := DecodeOrderEvent(EventPaymentFunded, topics, value)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["order_id"] != orderID || payload["buyer"] != buyer {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	amt, ok := payload["payment_amount"].(model.Amount)
	if !ok || amt.String() != paymentAmount.String() {
		t.Fatalf("expected exact payment_amount=%s, got %v", paymentAmount, payload["payment_amount"])
	}
}

func TestDecodeOrderEvent_WrongTopicCountFails(t *testing.T) {
	_, buyerVal := testAddressScVal(t)
	// OrderCreated expects 3 topic fields (order_id, buyer, distributor); give it 1.
	topics := []xdr.ScVal{testSymbolScVal(t, EventOrderCreated), buyerVal}
	_, err := DecodeOrderEvent(EventOrderCreated, topics, xdr.ScVal{})
	if err == nil {
		t.Fatal("expected a decode error for the wrong topic count, got nil")
	}
}

func TestDecodeOrderEvent_UnknownEventFails(t *testing.T) {
	_, err := DecodeOrderEvent("not_a_real_event", []xdr.ScVal{testSymbolScVal(t, "not_a_real_event")}, xdr.ScVal{})
	if err == nil {
		t.Fatal("expected a decode error for an unknown event name, got nil")
	}
}
