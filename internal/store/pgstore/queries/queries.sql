-- name: GetGames :many
SELECT * FROM games;

-- name: CreateNewGame :one
INSERT INTO games 
("state", "round", "created_at", "result", "deck_id")
VALUES 
(DEFAULT, DEFAULT, DEFAULT, DEFAULT, $1)
RETURNING *;

-- name: GetRoom :one
SELECT * FROM games
WHERE id=$1;

-- name: CreatePlayer :one
INSERT INTO players 
("name", "room_id")
VALUES
($1, $2)
RETURNING "id";

-- name: GetRoomPlayers :many
SELECT 
    "id" 
FROM players 
WHERE
    room_id=$1
ORDER BY ordem;



-- name: CreateMessage :one 
INSERT INTO chat_messages 
("room_id", "message", "player" )
VALUES 
($1, $2, $3)
RETURNING "id";

-- name: GetMessage :one
SELECT * FROM chat_messages
WHERE id=$1;

-- name: GetRoomMessages :many
SELECT * FROM chat_messages
WHERE room_id=$1;

-- name: SetRoomState :exec
UPDATE games 
SET 
"state"=$1
WHERE id=$2;


-- name: DeleteGameRoom :one
DELETE FROM games 
WHERE
    id=$1
RETURNING "id";

-- name: RemovePlayerFromRoom :one
DELETE FROM players 
WHERE id=$1
RETURNING "id";

-- name: GetAllRooms :many
SELECT * FROM games;


-- name: SetOrder :exec
UPDATE players 
SET "ordem"=$1
WHERE id=$2;