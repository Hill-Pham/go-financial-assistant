-- +goose Up
ALTER TABLE purchases
    ADD COLUMN kind TEXT NOT NULL DEFAULT 'EXPENSE'
        CHECK (kind IN ('EXPENSE', 'INCOME'));

ALTER TABLE purchases
    ADD CONSTRAINT chk_income_no_installment
        CHECK (kind != 'INCOME' OR type != 'INSTALLMENT');

CREATE INDEX idx_purchases_kind ON purchases(kind);

-- +goose Down
DROP INDEX IF EXISTS idx_purchases_kind;
ALTER TABLE purchases DROP CONSTRAINT IF EXISTS chk_income_no_installment;
ALTER TABLE purchases DROP COLUMN IF EXISTS kind;
