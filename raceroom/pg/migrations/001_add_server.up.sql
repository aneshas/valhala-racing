CREATE TABLE IF NOT EXISTS server
(
    id bigserial NOT NULL,
    user_email text NOT NULL,
    hours_reserved int NOT NULL,
    payment_received_at timestamptz NULL DEFAULT NULL,
    created_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (id)
);
