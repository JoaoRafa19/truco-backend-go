CREATE TYPE state AS ENUM ( 'truco', 'seis', 'nove' );

CREATE TABLE IF NOT EXISTS games (
    "id"            uuid    PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    "rodada"        INTEGER             NOT NULL DEFAULT 1,
    "created_at"    DATE                NOT NULL DEFAULT now(),
    "result"        JSONB,
    "players"       uuid[]              NOT NULL,
    "state"         state
);

---- create above / drop below ----

DROP TABLE IF EXISTS games;
DELETE TYPE STATE;