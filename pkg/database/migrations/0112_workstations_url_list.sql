-- +goose Up
CREATE TABLE workstations_url_lists (
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    nav_ident TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    url TEXT NOT NULL,
    duration INTERVAL NOT NULL DEFAULT '12 hours' CHECK (duration >= '1 hour' AND duration <= '12 hours'),
    description TEXT NOT NULL,
    PRIMARY KEY (id, nav_ident),
    UNIQUE (nav_ident, url)
);

INSERT INTO workstations_url_lists (nav_ident, created_at, expires_at, url, description)
WITH latest_history AS (
    SELECT DISTINCT ON (nav_ident)
    nav_ident,
    created_at,
    url_list
FROM workstations_url_list_history
ORDER BY nav_ident, created_at DESC
    )
SELECT DISTINCT
    h.nav_ident,
    h.created_at,
    h.created_at AS expired_at,
    TRIM(url_item) AS url,
    'Velg en beskrivelse for å kunne åpne' AS description
FROM latest_history h
    CROSS JOIN LATERAL unnest(string_to_array(h.url_list, E'\n')) AS url_item
WHERE TRIM(url_item) != '' AND TRIM(url_item) IS NOT NULL;

-- +goose Down
DROP TABLE workstations_url_lists;
