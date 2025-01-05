CREATE TABLE IF NOT EXISTS counters
(
    name       VARCHAR(255) PRIMARY KEY NOT NULL,
    value      INT                      NOT NULL,
    created_at TIMESTAMP                NOT NULL,
    updated_at TIMESTAMP                NOT NULL
);

CREATE TABLE IF NOT EXISTS gauges
(
    name       VARCHAR(255) PRIMARY KEY NOT NULL,
    value      DOUBLE PRECISION         NOT NULL,
    created_at TIMESTAMP                NOT NULL,
    updated_at TIMESTAMP                NOT NULL
);
