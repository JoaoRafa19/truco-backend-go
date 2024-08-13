-- Write your migrate up statements here
ALTER TABLE games DROP COLUMN rodada;
ALTER TABLE games ADD round INTEGER             NOT NULL DEFAULT 1;

---- create above / drop below ----
ALTER TABLE games ADD rodada INTEGER             NOT NULL DEFAULT 1;
ALTER TABLE games DROP COLUMN round;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
