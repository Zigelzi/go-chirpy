-- +goose Up
-- +goose StatementBegin
ALTER TABLE chirps
ADD deleted_at TIMESTAMPTZ;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE chirps
DROP deleted_at;

-- +goose StatementEnd
