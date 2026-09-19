-- Nexus indexed projection schema. PostgreSQL is a durable projection of
-- Stellar chain state, not the protocol source of truth — Stellar is.

CREATE TABLE indexer_checkpoints (
    id                  TEXT PRIMARY KEY,
    last_ledger         BIGINT NOT NULL,
    cursor              TEXT,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE stellar_events (
    event_id            TEXT PRIMARY KEY,
    ledger              BIGINT NOT NULL,
    ledger_closed_at    TIMESTAMPTZ NOT NULL,
    transaction_hash    TEXT NOT NULL,
    transaction_index   INTEGER NOT NULL,
    operation_index     INTEGER NOT NULL,
    event_type          TEXT NOT NULL,
    contract_id         TEXT NOT NULL,
    topics_xdr          TEXT NOT NULL,
    value_xdr           TEXT NOT NULL,
    decoded_payload      JSONB,
    decode_error        TEXT,
    first_observed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE assets (
    asset               TEXT PRIMARY KEY,
    issuer              TEXT NOT NULL,
    active              BOOLEAN NOT NULL,
    first_seen_ledger   BIGINT NOT NULL,
    last_updated_ledger BIGINT NOT NULL
);

CREATE TABLE distributions (
    asset                   TEXT NOT NULL REFERENCES assets(asset),
    distributor             TEXT NOT NULL,
    eligibility_authority   TEXT NOT NULL,
    active                  BOOLEAN NOT NULL,
    first_seen_ledger       BIGINT NOT NULL,
    last_updated_ledger     BIGINT NOT NULL,
    PRIMARY KEY (asset, distributor)
);

CREATE TABLE eligibilities (
    asset               TEXT NOT NULL,
    distributor         TEXT NOT NULL,
    buyer               TEXT NOT NULL,
    valid_until_ledger  BIGINT NOT NULL,
    active              BOOLEAN NOT NULL,
    last_updated_ledger BIGINT NOT NULL,
    PRIMARY KEY (asset, distributor, buyer),
    FOREIGN KEY (asset, distributor) REFERENCES distributions(asset, distributor)
);

CREATE TABLE orders (
    order_id                BIGINT PRIMARY KEY,
    buyer                   TEXT NOT NULL,
    distributor             TEXT NOT NULL,
    asset                   TEXT NOT NULL,
    payment_asset           TEXT NOT NULL,
    -- Exact i128 token amounts. NUMERIC(39,0) comfortably covers the full
    -- i128 range (max ~1.7e38, 39 digits) with no fractional component and
    -- no floating-point rounding.
    asset_amount            NUMERIC(39,0) NOT NULL,
    payment_amount          NUMERIC(39,0) NOT NULL,
    created_at_ledger       BIGINT NOT NULL,
    expires_at_ledger       BIGINT NOT NULL,
    status                  TEXT NOT NULL,
    payment_funded          BOOLEAN NOT NULL DEFAULT false,
    asset_funded            BOOLEAN NOT NULL DEFAULT false,
    created_transaction_hash TEXT NOT NULL,
    last_transaction_hash   TEXT NOT NULL,
    last_updated_ledger     BIGINT NOT NULL
);

CREATE TABLE order_events (
    id                  BIGSERIAL PRIMARY KEY,
    order_id            BIGINT NOT NULL REFERENCES orders(order_id),
    event_type          TEXT NOT NULL,
    event_id            TEXT NOT NULL REFERENCES stellar_events(event_id),
    ledger              BIGINT NOT NULL,
    transaction_hash    TEXT NOT NULL,
    payload             JSONB,
    UNIQUE (order_id, event_id)
);

CREATE TABLE transactions (
    transaction_hash    TEXT PRIMARY KEY,
    ledger              BIGINT,
    status              TEXT NOT NULL,
    first_observed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_observed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    source_contract     TEXT,
    order_id            BIGINT REFERENCES orders(order_id)
);
