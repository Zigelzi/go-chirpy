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

-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = $1;

-- name: UpdateUserCredentials :one
UPDATE users
SET
    email = $2,
    hashed_password = $3,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    *;