-- +goose Up
CREATE TABLE accounts(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id text NOT NULL,
    balance bigint NOT NULL DEFAULT 0 CHECK(balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE transactions(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES accounts(id),
    type text NOT NULL CHECK(type IN ('deposit', 'withdraw')),
    amount bigint NOT NULL CHECK(amount > 0),
    balance_after bigint NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX transactions_account_created_idx ON transactions (account_id, created_at DESC);
-- +goose Down
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;