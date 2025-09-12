-- name: CreateUser :one
INSERT INTO
    users (
        id,
        email,
        hashed_password,
        created_at,
        updated_at
    )
VALUES
    ($1, $2, $3, NOW(), NOW())
RETURNING
    *;

-- name: ResetUsers :execrows
DELETE FROM users;

-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    email = $1;