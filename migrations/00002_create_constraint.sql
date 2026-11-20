-- +goose Up
ALTER TABLE accounts ADD CONSTRAINT accounts_owner_id_key UNIQUE (owner_id);
-- +goose Down
ALTER TABLE accounts DROP CONSTRAINT accounts_owner_id_key;
