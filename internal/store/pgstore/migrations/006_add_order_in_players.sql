-- Write your migrate up statements here
ALTER TABLE players ADD ordem INTEGER NOT NULL DEFAULT -1;

---- create above / drop below ----
ALTER TABLE DROP COLUMN ordem;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
