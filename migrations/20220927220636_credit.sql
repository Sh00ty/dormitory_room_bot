-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS credits(
    channel_id bigint,
    user_id text,
    credit bigint,
    PRIMARY KEY (channel_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS credit;
-- +goose StatementEnd
