package indexing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// Registry event names, exactly as emitted by the #[contractevent] macro
// (the Rust struct name lowercased to snake_case) — verified against the
// real generated Registry ABI (packages/contracts/registry/src/index.ts),
// not guessed.
const (
	EventAssetRegistered        = "asset_registered"
	EventAssetDeactivated       = "asset_deactivated"
	EventDistributionRegistered = "distribution_registered"
	EventDistributionRevoked    = "distribution_revoked"
	EventEligibilitySet         = "eligibility_set"
	EventEligibilityRevoked     = "eligibility_revoked"
	EventAdminTransferProposed  = "admin_transfer_proposed"
	EventAdminTransferred       = "admin_transferred"
	EventAdminTransferCancelled = "admin_transfer_cancelled"
)

// dataFields extracts named non-topic ("data") fields from an event's value
// ScVal. Soroban's #[contractevent] macro defaults to data_format = "map"
// regardless of field count — including zero and one-field events — so the
// value is always a Map<Symbol, Val> keyed by field name. This is not
// documentation-inferred: it was confirmed by decoding the real compiled
// ScSpecEventV0.DataFormat for every event in both the Registry and Order
// ABIs (all report Map), so there is no single-value or zero-field special
// case to handle here.
func dataFields(value xdr.ScVal, names ...string) (map[string]xdr.ScVal, error) {
	if len(names) == 0 {
		return map[string]xdr.ScVal{}, nil
	}

	m, ok := value.GetMap()
	if !ok || m == nil {
		return nil, fmt.Errorf("expected a Map value for a %d-field event", len(names))
	}
	result := make(map[string]xdr.ScVal, len(names))
	for _, entry := range *m {
		key, err := scValToSymbol(entry.Key)
		if err != nil {
			continue
		}
		result[key] = entry.Val
	}
	for _, name := range names {
		if _, ok := result[name]; !ok {
			return nil, fmt.Errorf("missing expected field %q in event data map", name)
		}
	}
	return result, nil
}

// DecodeRegistryEvent decodes a raw Registry event into a JSON-serializable
// payload keyed by field name (topics and data merged), for storage in
// stellar_events.decoded_payload.
func DecodeRegistryEvent(eventName string, topics []xdr.ScVal, value xdr.ScVal) (map[string]any, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("event has no topics")
	}
	// topics[0] is the event name Symbol; #[topic] fields follow.
	topicArgs := topics[1:]

	switch eventName {
	case EventAssetRegistered:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		asset, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "issuer")
		if err != nil {
			return nil, err
		}
		issuer, err := scValToAddress(fields["issuer"])
		if err != nil {
			return nil, err
		}
		return map[string]any{"asset": asset, "issuer": issuer}, nil

	case EventAssetDeactivated:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		asset, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		return map[string]any{"asset": asset}, nil

	case EventDistributionRegistered:
		if len(topicArgs) != 2 {
			return nil, fmt.Errorf("%s: expected 2 topic fields, got %d", eventName, len(topicArgs))
		}
		asset, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		distributor, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "eligibility_authority")
		if err != nil {
			return nil, err
		}
		eligibilityAuthority, err := scValToAddress(fields["eligibility_authority"])
		if err != nil {
			return nil, err
		}
		return map[string]any{"asset": asset, "distributor": distributor, "eligibility_authority": eligibilityAuthority}, nil

	case EventDistributionRevoked:
		if len(topicArgs) != 2 {
			return nil, fmt.Errorf("%s: expected 2 topic fields, got %d", eventName, len(topicArgs))
		}
		asset, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		distributor, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		return map[string]any{"asset": asset, "distributor": distributor}, nil

	case EventEligibilitySet:
		if len(topicArgs) != 3 {
			return nil, fmt.Errorf("%s: expected 3 topic fields, got %d", eventName, len(topicArgs))
		}
		asset, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		distributor, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		buyer, err := scValToAddress(topicArgs[2])
		if err != nil {
			return nil, err
		}
		fields, err := dataFields(value, "valid_until_ledger")
		if err != nil {
			return nil, err
		}
		validUntilLedger, err := scValToU32(fields["valid_until_ledger"])
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"asset": asset, "distributor": distributor, "buyer": buyer,
			"valid_until_ledger": validUntilLedger,
		}, nil

	case EventEligibilityRevoked:
		if len(topicArgs) != 3 {
			return nil, fmt.Errorf("%s: expected 3 topic fields, got %d", eventName, len(topicArgs))
		}
		asset, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		distributor, err := scValToAddress(topicArgs[1])
		if err != nil {
			return nil, err
		}
		buyer, err := scValToAddress(topicArgs[2])
		if err != nil {
			return nil, err
		}
		return map[string]any{"asset": asset, "distributor": distributor, "buyer": buyer}, nil

	case EventAdminTransferProposed:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		newAdmin, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		return map[string]any{"new_admin": newAdmin}, nil

	case EventAdminTransferred:
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

	case EventAdminTransferCancelled:
		if len(topicArgs) != 1 {
			return nil, fmt.Errorf("%s: expected 1 topic field, got %d", eventName, len(topicArgs))
		}
		pendingAdmin, err := scValToAddress(topicArgs[0])
		if err != nil {
			return nil, err
		}
		return map[string]any{"pending_admin": pendingAdmin}, nil

	default:
		return nil, fmt.Errorf("unknown Registry event type %q", eventName)
	}
}

// ProjectRegistryEvent applies a decoded Registry event to the assets,
// distributions, and eligibilities projection tables. Admin events have no
// dedicated projection table — they are recorded in stellar_events'
// decoded_payload only, for audit purposes.
func ProjectRegistryEvent(ctx context.Context, tx *sql.Tx, eventName string, payload map[string]any, ledger int64) error {
	switch eventName {
	case EventAssetRegistered:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO assets (asset, issuer, active, first_seen_ledger, last_updated_ledger)
			VALUES ($1, $2, true, $3, $3)
			ON CONFLICT (asset) DO UPDATE SET
				issuer = EXCLUDED.issuer, active = true, last_updated_ledger = $3
			WHERE assets.last_updated_ledger <= $3`,
			payload["asset"], payload["issuer"], ledger)
		return err

	case EventAssetDeactivated:
		_, err := tx.ExecContext(ctx, `
			UPDATE assets SET active = false, last_updated_ledger = $2
			WHERE asset = $1 AND last_updated_ledger <= $2`,
			payload["asset"], ledger)
		return err

	case EventDistributionRegistered:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO distributions (asset, distributor, eligibility_authority, active, first_seen_ledger, last_updated_ledger)
			VALUES ($1, $2, $3, true, $4, $4)
			ON CONFLICT (asset, distributor) DO UPDATE SET
				eligibility_authority = EXCLUDED.eligibility_authority, active = true, last_updated_ledger = $4
			WHERE distributions.last_updated_ledger <= $4`,
			payload["asset"], payload["distributor"], payload["eligibility_authority"], ledger)
		return err

	case EventDistributionRevoked:
		_, err := tx.ExecContext(ctx, `
			UPDATE distributions SET active = false, last_updated_ledger = $3
			WHERE asset = $1 AND distributor = $2 AND last_updated_ledger <= $3`,
			payload["asset"], payload["distributor"], ledger)
		return err

	case EventEligibilitySet:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO eligibilities (asset, distributor, buyer, valid_until_ledger, active, last_updated_ledger)
			VALUES ($1, $2, $3, $4, true, $5)
			ON CONFLICT (asset, distributor, buyer) DO UPDATE SET
				valid_until_ledger = EXCLUDED.valid_until_ledger, active = true, last_updated_ledger = $5
			WHERE eligibilities.last_updated_ledger <= $5`,
			payload["asset"], payload["distributor"], payload["buyer"], payload["valid_until_ledger"], ledger)
		return err

	case EventEligibilityRevoked:
		_, err := tx.ExecContext(ctx, `
			UPDATE eligibilities SET active = false, last_updated_ledger = $4
			WHERE asset = $1 AND distributor = $2 AND buyer = $3 AND last_updated_ledger <= $4`,
			payload["asset"], payload["distributor"], payload["buyer"], ledger)
		return err

	case EventAdminTransferProposed, EventAdminTransferred, EventAdminTransferCancelled:
		return nil // recorded via decoded_payload only; no dedicated projection table

	default:
		return fmt.Errorf("unknown Registry event type %q", eventName)
	}
}

// marshalPayload is a small helper so callers can store a decoded event's
// payload as JSON without repeating error handling.
func marshalPayload(payload map[string]any) ([]byte, error) {
	return json.Marshal(payload)
}
