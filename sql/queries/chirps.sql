-- name: CreateChirpy :one
INSERT INTO chirps (id, body, user_id, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpyByID :one
SELECT * FROM chirps
WHERE id = $1 LIMIT 1;
