package indexing

import (
	"context"
	"database/sql"
	"fmt"

	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func decodeTopicsXDR(topicXDR []string) ([]xdr.ScVal, error) {
	topics := make([]xdr.ScVal, 0, len(topicXDR))
	for _, t := range topicXDR {
		var v xdr.ScVal
		if err := xdr.SafeUnmarshalBase64(t, &v); err != nil {
			return nil, fmt.Errorf("decode topic xdr: %w", err)
		}
		topics = append(topics, v)
	}
	return topics, nil
}

func decodeValueXDR(valueXDR string) (xdr.ScVal, error) {
	var v xdr.ScVal
	if valueXDR == "" {
		return v, nil
	}
	if err := xdr.SafeUnmarshalBase64(valueXDR, &v); err != nil {
		return v, fmt.Errorf("decode value xdr: %w", err)
	}
	return v, nil
}

// decodedEvent is a successfully decoded event, not yet projected.
type decodedEvent struct {
	name       string
	payload    map[string]any
	isRegistry bool
}

// decodeEvent decodes one raw event's topics/value, dispatching by which
// configured contract emitted it. This is pure decoding with no database
// access — callers must persist the raw stellar_events row before calling
// projectEvent, since order_events has a foreign key on stellar_events
// (event_id) that a projection insert would otherwise violate.
//
// Returns (nil, nil) for events from a contract this indexer isn't
// configured to decode. A non-nil error means decoding failed — the caller
// should record it as decode_error and continue with the rest of the batch,
// per spec §16.3 (never discard the raw event, never let one bad event
// abort the batch).
func decodeEvent(event protocol.EventInfo, registryContractID, orderContractID string) (*decodedEvent, error) {
	isRegistry := registryContractID != "" && event.ContractID == registryContractID
	isOrder := orderContractID != "" && event.ContractID == orderContractID
	if !isRegistry && !isOrder {
		return nil, nil
	}

	topics, err := decodeTopicsXDR(event.TopicXDR)
	if err != nil {
		return nil, err
	}
	if len(topics) == 0 {
		return nil, fmt.Errorf("event has no topics")
	}

	eventName, err := scValToSymbol(topics[0])
	if err != nil {
		return nil, fmt.Errorf("decode event name: %w", err)
	}

	value, err := decodeValueXDR(event.ValueXDR)
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	if isRegistry {
		payload, err = DecodeRegistryEvent(eventName, topics, value)
	} else {
		payload, err = DecodeOrderEvent(eventName, topics, value)
	}
	if err != nil {
		return nil, err
	}

	return &decodedEvent{name: eventName, payload: payload, isRegistry: isRegistry}, nil
}

// projectEvent applies a successfully decoded event's projection update.
// Must only be called after the corresponding stellar_events row has been
// inserted in the same transaction (see decodeEvent's doc comment).
func projectEvent(ctx context.Context, tx *sql.Tx, decoded *decodedEvent, event protocol.EventInfo) error {
	if decoded.isRegistry {
		if err := ProjectRegistryEvent(ctx, tx, decoded.name, decoded.payload, int64(event.Ledger)); err != nil {
			return fmt.Errorf("project %s: %w", decoded.name, err)
		}
		return nil
	}
	if err := ProjectOrderEvent(ctx, tx, decoded.name, decoded.payload, int64(event.Ledger), event.TransactionHash, event.ID); err != nil {
		return fmt.Errorf("project %s: %w", decoded.name, err)
	}
	return nil
}
