-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ALTER COLUMN updated_at TYPE TIMESTAMPTZ,
ALTER COLUMN created_at TYPE TIMESTAMPTZ;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
ALTER COLUMN updated_at TYPE TIMESTAMP,
ALTER COLUMN created_at TYPE TIMESTAMP;

-- +goose StatementEnd
