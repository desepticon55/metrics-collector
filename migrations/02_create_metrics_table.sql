-- +goose Up
CREATE TABLE mt_cl.metrics
(
    name       VARCHAR(50),
    type       VARCHAR(20),
    value      TEXT        NOT NULL,
    PRIMARY KEY (name, type)
);

-- +goose Down
DROP TABLE mt_cl.metrics;