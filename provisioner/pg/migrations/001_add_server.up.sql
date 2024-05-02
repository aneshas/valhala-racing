CREATE TABLE IF NOT EXISTS provisioned_server
(
    id bigserial NOT NULL,
    server_id int NOT NULL,
    expires_at timestamptz NOT NULL,
    termination_scheduled_at timestamptz NULL DEFAULT NULL,
    terminated_at timestamptz NULL DEFAULT NULL,
    created_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (id)
);
