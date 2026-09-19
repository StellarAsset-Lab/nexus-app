package indexing

import (
	"context"
	"database/sql"
	"fmt"

	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// projectionFailure distinguishes a real infrastructure/database error
// (which must abort and retry the whole batch transaction) from an event
// decode failure (which is recorded per-event via decode_error and must not
// block the rest of the batch — see spec §16.3).
type projectionFailure struct{ err error }

func (p *projectionFailure) Error() string { return p.err.Error() }
func (p *projectionFailure) Unwrap() error { return p.err }

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

// decodeAndProjectEvent decodes one raw event and applies its projection
// update within tx, dispatching by which configured contract emitted it.
// Returns (nil, nil) for events from contracts this indexer isn't
// configured to decode (e.g. Order events before commit 16 adds Order
// decoding). A plain returned error means decoding failed (caller records
// decode_error and continues); a *projectionFailure means a real database
// error occurred applying the projection (caller must abort the batch).
func decodeAndProjectEvent(ctx context.Context, tx *sql.Tx, event protocol.EventInfo, registryContractID string) (map[string]any, error) {
	if registryContractID == "" || event.ContractID != registryContractID {
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

	payload, err := DecodeRegistryEvent(eventName, topics, value)
	if err != nil {
		return nil, err
	}

	if err := ProjectRegistryEvent(ctx, tx, eventName, payload, int64(event.Ledger)); err != nil {
		return nil, &projectionFailure{err: fmt.Errorf("project %s: %w", eventName, err)}
	}

	return payload, nil
}
