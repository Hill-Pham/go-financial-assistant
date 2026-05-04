-- +goose Up
ALTER TABLE purchases
    ADD COLUMN IF NOT EXISTS transfer_direction TEXT
        CHECK (transfer_direction IN ('IN', 'OUT'));

-- +goose Down
ALTER TABLE purchases DROP COLUMN IF EXISTS transfer_direction;
