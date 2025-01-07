CREATE TABLE IF NOT EXISTS counters
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    value      INT          NOT NULL,
    created_at TIMESTAMP    NOT NULL,
    updated_at TIMESTAMP    NOT NULL
);

CREATE TABLE IF NOT EXISTS gauges
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255)     NOT NULL,
    value      DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP        NOT NULL,
    updated_at TIMESTAMP        NOT NULL
);
