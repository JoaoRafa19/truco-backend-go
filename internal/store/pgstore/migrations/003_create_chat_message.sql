-- Write your migrate up statements here
CREATE TABLE IF NOT EXISTS  chat_messages (
    "id"            uuid        PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    "room_id"       uuid                    NOT NULL,
    "message"       VARCHAR                 NOT NULL,
    "player"        uuid                    NOT NULL,
    "created_at"    DATE                    NOT NULL DEFAULT now(),

    FOREIGN KEY (room_id)   REFERENCES game(id) ON DELETE CASCADE,
    FOREIGN KEY (player)    REFERENCES players(id)
);

---- create above / drop below ----
DROP TABLE IF EXISTS chat_messages ;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
