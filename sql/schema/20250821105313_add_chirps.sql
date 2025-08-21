-- +goose Up
-- +goose StatementBegin
CREATE TABLE chirps (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    body TEXT NOT NULL,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE NOT NULL
)
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE chirps;

-- +goose StatementEnd
