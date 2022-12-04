-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks (
    id text,
    title text NOT NULL,
    notify_time timestamp,
    notify_interval bigint,
    workers text[],
    worker_count bigint,
    first_worker bigint,
    channel_id bigint,
    author text,
    type int,
    stage int,
    primary key (channel_id, id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS task;
-- +goose StatementEnd
