package indexing

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

// Order event names, exactly as emitted by the #[contractevent] macro —
// verified against the real generated Order ABI
// (packages/contracts/order/src/index.ts).
const (
	EventOrderCreated                = "order_created"
	EventPaymentFunded               = "payment_funded"
	EventAssetFunded                 = "asset_funded"
	EventOrderSettled                = "order_settled"
	EventOrderCancelled              = "order_cancelled"
	EventOrderExpired                = "order_expired"
	EventGatewayPaused               = "gateway_paused"
	EventGatewayUnpaused             = "gateway_unpaused"
	EventOrderAdminTransferProposed  = "admin_transfer_proposed"
	EventOrderAdminTransferred       = "admin_transferred"
	EventOrderAdminTransferCancelled = "admin_transfer_cancelled"
)

// DecodeOrderEvent decodes a raw Order event into a JSON-serializable
// payload keyed by field name (topics and data merged).
func DecodeOrderEvent(eventName string, topics []xdr.ScVal, value xdr.ScVal) (map[string]any, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("event has no topics")
	}
	topicArgs := topics[1:]

	switch eventName {
	case EventOrderCreated:
		if len(topicArgs) != 3 {
			return nil, fmt.Errorf("%s: expected 3 topic fields, got %d", eventName, len(topicArgs))
		}
		orderID, err := scValToU64(topicArgs[0])
		if err != nil {
			return nil, err
		}
		buyer, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		distributor, err := scValToAddress(topicArgs[2])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "asset", "payment_asset", "asset_amount", "payment_amount", "expires_at_ledger")
		if err != nil {
			return nil, err
		}
		asset, err := scValToAddress(fields["asset"])
		if err != nil {
			return nil, err
		}
		paymentAsset, err := scValToAddress(fields["payment_asset"])
		if err != nil {
			return nil, err
		}
		assetAmount, err := scValToAmount(fields["asset_amount"])
		if err != nil {
			return nil, err
		}
		paymentAmount, err := scValToAmount(fields["payment_amount"])
		if err != nil {
			return nil, err
		}
		expiresAtLedger, err := scValToU32(fields["expires_at_ledger"])
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"order_id": orderID, "buyer": buyer, "distributor": distributor,
			"asset": asset, "payment_asset": paymentAsset,
			"asset_amount": assetAmount, "payment_amount": paymentAmount,
			"expires_at_ledger": expiresAtLedger,
		}, nil

	case EventPaymentFunded:
		if len(topicArgs) != 2 {
			return nil, fmt.Errorf("%s: expected 2 topic fields, got %d", eventName, len(topicArgs))
		}
		orderID, err := scValToU64(topicArgs[0])
		if err != nil {
			return nil, err
		}
		buyer, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "payment_amount")
		if err != nil {
			return nil, err
		}
		paymentAmount, err := scValToAmount(fields["payment_amount"])
		if err != nil {
			return nil, err
		}
		return map[string]any{"order_id": orderID, "buyer": buyer, "payment_amount": paymentAmount}, nil

	case EventAssetFunded:
		if len(topicArgs) != 2 {
			return nil, fmt.Errorf("%s: expected 2 topic fields, got %d", eventName, len(topicArgs))
		}
		orderID, err := scValToU64(topicArgs[0])
		if err != nil {
			return nil, err
		}
		distributor, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "asset_amount")
		if err != nil {
			return nil, err
		}
		assetAmount, err := scValToAmount(fields["asset_amount"])
		if err != nil {
			return nil, err
		}
		return map[string]any{"order_id": orderID, "distributor": distributor, "asset_amount": assetAmount}, nil

	case EventOrderSettled:
		if len(topicArgs) != 3 {
			return nil, fmt.Errorf("%s: expected 3 topic fields, got %d", eventName, len(topicArgs))
		}
		orderID, err := scValToU64(topicArgs[0])
		if err != nil {
			return nil, err
		}
		buyer, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		distributor, err := scValToAddress(topicArgs[2])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "asset_amount", "payment_amount")
		if err != nil {
			return nil, err
		}
		assetAmount, err := scValToAmount(fields["asset_amount"])
		if err != nil {
			return nil, err
		}
		paymentAmount, err := scValToAmount(fields["payment_amount"])
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"order_id": orderID, "buyer": buyer, "distributor": distributor,
			"asset_amount": assetAmount, "payment_amount": paymentAmount,
		}, nil

	case EventOrderCancelled:
		if len(topicArgs) != 2 {
			return nil, fmt.Errorf("%s: expected 2 topic fields, got %d", eventName, len(topicArgs))
		}
		orderID, err := scValToU64(topicArgs[0])
		if err != nil {
			return nil, err
		}
		cancelledBy, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		return map[string]any{"order_id": orderID, "cancelled_by": cancelledBy}, nil

	case EventOrderExpired:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		orderID, err := scValToU64(topicArgs[0])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "payment_refunded", "asset_refunded")
		if err != nil {
			return nil, err
		}
		paymentRefunded, err := scValToBool(fields["payment_refunded"])
		if err != nil {
			return nil, err
		}
		assetRefunded, err := scValToBool(fields["asset_refunded"])
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"order_id": orderID, "payment_refunded": paymentRefunded, "asset_refunded": assetRefunded,
		}, nil

	case EventGatewayPaused, EventGatewayUnpaused:
		return map[string]any{}, nil

	case EventOrderAdminTransferProposed:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		newAdmin, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		return map[string]any{"new_admin": newAdmin}, nil

	case EventOrderAdminTransferred:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		previousAdmin, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "new_admin")
		if err != nil {
			return nil, err
		}
		newAdmin, err := scValToAddress(fields["new_admin"])
		if err != nil {
			return nil, err
		}
		return map[string]any{"previous_admin": previousAdmin, "new_admin": newAdmin}, nil

	case EventOrderAdminTransferCancelled:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		pendingAdmin, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		return map[string]any{"pending_admin": pendingAdmin}, nil

	default:
		return nil, fmt.Errorf("unknown Order event type %q", eventName)
	}
}

// orderTerminalStatuses are never moved backward or sideways by a later
// duplicate or malformed event — see spec §16.4.
const orderStatusClause = `status NOT IN ('Settled', 'Cancelled', 'Expired')`

// ProjectOrderEvent applies a decoded Order event to the orders and
// order_events projection tables. Gateway pause/unpause and admin events
// have no order_id and are not linked into order_events.
func ProjectOrderEvent(
	ctx context.Context, tx *sql.Tx, eventName string, payload map[string]any,
	ledger int64, txHash, eventID string,
) error {
	switch eventName {
	case EventOrderCreated:
		orderID := payload["order_id"]
		_, err := tx.ExecContext(ctx, `
			INSERT INTO orders (
				order_id, buyer, distributor, asset, payment_asset, asset_amount, payment_amount,
				created_at_ledger, expires_at_ledger, status, payment_funded, asset_funded,
				created_transaction_hash, last_transaction_hash, last_updated_ledger
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'Created', false, false, $10, $10, $8)
			ON CONFLICT (order_id) DO NOTHING`,
			orderID, payload["buyer"], payload["distributor"], payload["asset"], payload["payment_asset"],
			amountValue(payload["asset_amount"]), amountValue(payload["payment_amount"]),
			ledger, payload["expires_at_ledger"], txHash,
		)
		if err != nil {
			return err
		}
		return insertOrderEvent(ctx, tx, orderID, eventName, eventID, ledger, txHash, payload)

	case EventPaymentFunded:
		orderID := payload["order_id"]
		_, err := tx.ExecContext(ctx, `
			UPDATE orders SET payment_funded = true, last_transaction_hash = $2, last_updated_ledger = $3
			WHERE order_id = $1 AND last_updated_ledger <= $3 AND `+orderStatusClause,
			orderID, txHash, ledger)
		if err != nil {
			return err
		}
		return insertOrderEvent(ctx, tx, orderID, eventName, eventID, ledger, txHash, payload)

	case EventAssetFunded:
		orderID := payload["order_id"]
		_, err := tx.ExecContext(ctx, `
			UPDATE orders SET asset_funded = true, last_transaction_hash = $2, last_updated_ledger = $3
			WHERE order_id = $1 AND last_updated_ledger <= $3 AND `+orderStatusClause,
			orderID, txHash, ledger)
		if err != nil {
			return err
		}
		return insertOrderEvent(ctx, tx, orderID, eventName, eventID, ledger, txHash, payload)

	case EventOrderSettled:
		orderID := payload["order_id"]
		_, err := tx.ExecContext(ctx, `
			UPDATE orders SET status = 'Settled', last_transaction_hash = $2, last_updated_ledger = $3
			WHERE order_id = $1 AND last_updated_ledger <= $3 AND `+orderStatusClause,
			orderID, txHash, ledger)
		if err != nil {
			return err
		}
		return insertOrderEvent(ctx, tx, orderID, eventName, eventID, ledger, txHash, payload)

	case EventOrderCancelled:
		orderID := payload["order_id"]
		_, err := tx.ExecContext(ctx, `
			UPDATE orders SET status = 'Cancelled', last_transaction_hash = $2, last_updated_ledger = $3
			WHERE order_id = $1 AND last_updated_ledger <= $3 AND `+orderStatusClause,
			orderID, txHash, ledger)
		if err != nil {
			return err
		}
		return insertOrderEvent(ctx, tx, orderID, eventName, eventID, ledger, txHash, payload)

	case EventOrderExpired:
		orderID := payload["order_id"]
		_, err := tx.ExecContext(ctx, `
			UPDATE orders SET status = 'Expired', last_transaction_hash = $2, last_updated_ledger = $3
			WHERE order_id = $1 AND last_updated_ledger <= $3 AND `+orderStatusClause,
			orderID, txHash, ledger)
		if err != nil {
			return err
		}
		return insertOrderEvent(ctx, tx, orderID, eventName, eventID, ledger, txHash, payload)

	case EventGatewayPaused, EventGatewayUnpaused,
		EventOrderAdminTransferProposed, EventOrderAdminTransferred, EventOrderAdminTransferCancelled:
		return nil // no order_id; recorded via decoded_payload only

	default:
		return fmt.Errorf("unknown Order event type %q", eventName)
	}
}

func amountValue(v any) string {
	if a, ok := v.(model.Amount); ok {
		return a.String()
	}
	return "0"
}

func insertOrderEvent(ctx context.Context, tx *sql.Tx, orderID any, eventName, eventID string, ledger int64, txHash string, payload map[string]any) error {
	payloadBytes, err := marshalPayload(payload)
	if err != nil {
		return fmt.Errorf("marshal order event payload: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO order_events (order_id, event_type, event_id, ledger, transaction_hash, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (order_id, event_id) DO NOTHING`,
		orderID, eventName, eventID, ledger, txHash, payloadBytes)
	return err
}
