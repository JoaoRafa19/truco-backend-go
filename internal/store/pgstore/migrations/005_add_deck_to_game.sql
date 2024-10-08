-- Write your migrate up statements here

ALTER TABLE games 
ADD deck_id VARCHAR(255) UNIQUE NOT NULL;

CREATE INDEX idx_salas_deckid ON games (deck_id);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_salas_deckid;

ALTER TABLE games
DROP COLUMN deck_id;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
