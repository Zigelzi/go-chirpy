-- name: CreateChirp :one
INSERT INTO
    chirps (id, body, user_id, created_at, updated_at)
VALUES
    ($1, $2, $3, NOW(), NOW())
RETURNING
    *;

-- name: GetChirps :many
SELECT
    *
FROM
    chirps
WHERE
    deleted_at is null
ORDER BY
    created_at asc;

-- name: GetChirpsByAuthorId :many
SELECT
    *
FROM
    chirps
WHERE
    deleted_at is null
    AND user_id = $1
ORDER BY
    created_at asc;

-- name: GetChirp :one
SELECT
    *
FROM
    chirps
WHERE
    id = $1
    and deleted_at is null;

-- name: DeleteChirp :one
UPDATE chirps
SET
    deleted_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1
    and deleted_at is null
RETURNING
    *;