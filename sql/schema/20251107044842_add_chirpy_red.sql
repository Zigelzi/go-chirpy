-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD is_chirpy_red BOOLEAN NOT NULL DEFAULT false;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP is_chirpy_red;

-- +goose StatementEnd
