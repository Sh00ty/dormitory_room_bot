-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id text PRIMARY KEY,
    phone_number text,
    username text NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
