-- +goose Up

-- River pro migration 004 [up]
CREATE TABLE river_job_dead_letter (
    -- laid out identically to the physical layout of `river_job`
                                       id bigserial PRIMARY KEY,
                                       state river_job_state NOT NULL DEFAULT 'available',
                                       attempt smallint NOT NULL,
                                       max_attempts smallint NOT NULL,
                                       attempted_at timestamptz,
                                       created_at timestamptz NOT NULL,
                                       finalized_at timestamptz,
                                       scheduled_at timestamptz NOT NULL,
                                       priority smallint NOT NULL,
                                       args jsonb NOT NULL,
                                       attempted_by text[],
                                       errors jsonb[],
                                       kind text NOT NULL,
                                       metadata jsonb NOT NULL,
                                       queue text NOT NULL,
                                       tags varchar(255)[] NOT NULL,
                                       unique_key bytea,
                                       unique_states bit(8),
                                       partition_key text
);

-- +goose Down

-- River pro migration 004 [down]
DROP TABLE river_job_dead_letter;
