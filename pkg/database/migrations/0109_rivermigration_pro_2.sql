-- +goose Up

-- River pro migration 002 [up]
CREATE UNLOGGED TABLE /* TEMPLATE: schema */river_producer (
                                                               id bigserial PRIMARY KEY,
                                                               client_id text NOT NULL,
                                                               queue_name text NOT NULL,

                                                               max_workers int NOT NULL CHECK (max_workers >= 0),
                                                               metadata jsonb NOT NULL DEFAULT '{}',
                                                               paused_at timestamptz,
                                                               created_at timestamptz NOT NULL DEFAULT now(),
                                                               updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down

-- River pro migration 002 [down]
DROP TABLE /* TEMPLATE: schema */river_producer;
