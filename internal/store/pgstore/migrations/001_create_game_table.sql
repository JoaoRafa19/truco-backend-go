DROP TYPE IF EXISTS state;
CREATE TYPE state AS ENUM ('normal', 'truco', 'seis', 'nove');

CREATE TABLE IF NOT EXISTS games (
    "id"            uuid    PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    "rodada"        INTEGER             NOT NULL DEFAULT 1,
    "created_at"    TIMESTAMP           NOT NULL DEFAULT now(),
    "result"        JSONB,
    "state"         state               NOT NULL DEFAULT 'normal'::state
);

---- create above / drop below ----

DROP TABLE IF EXISTS games;
DROP TYPE IF EXISTS state;