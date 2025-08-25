CREATE TABLE IF NOT EXISTS counters
(
    name       VARCHAR(255) PRIMARY KEY,
    value      BIGINT    NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS gauges
(
    name       VARCHAR(255) PRIMARY KEY,
    value      DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP        NOT NULL
);
