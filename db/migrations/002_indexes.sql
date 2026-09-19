-- Deterministic pagination and common filter paths. Every list endpoint
-- orders by ledger (or event id) plus a unique tiebreaker — never by
-- database default row order.

CREATE INDEX idx_stellar_events_ledger ON stellar_events (ledger, event_id);
CREATE INDEX idx_stellar_events_contract ON stellar_events (contract_id, ledger);
CREATE INDEX idx_stellar_events_type ON stellar_events (event_type, ledger);
CREATE INDEX idx_stellar_events_decode_error ON stellar_events (ledger) WHERE decode_error IS NOT NULL;

CREATE INDEX idx_assets_active ON assets (active);

CREATE INDEX idx_distributions_distributor ON distributions (distributor);

CREATE INDEX idx_eligibilities_buyer ON eligibilities (buyer);

CREATE INDEX idx_orders_status ON orders (status, created_at_ledger DESC, order_id DESC);
CREATE INDEX idx_orders_asset ON orders (asset, created_at_ledger DESC, order_id DESC);
CREATE INDEX idx_orders_distributor ON orders (distributor, created_at_ledger DESC, order_id DESC);
CREATE INDEX idx_orders_buyer ON orders (buyer, created_at_ledger DESC, order_id DESC);

CREATE INDEX idx_order_events_order ON order_events (order_id, ledger, id);

CREATE INDEX idx_transactions_order ON transactions (order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_transactions_status ON transactions (status);
