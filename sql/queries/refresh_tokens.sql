-- name: CreateRefreshToken :exec
INSERT INTO
    refresh_tokens (
        token,
        expires_at,
        user_id,
        created_at,
        updated_at
    )
VALUES
    ($1, $2, $3, NOW(), NOW());

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
    updated_at = NOW(),
    revoked_at = NOW()
WHERE
    token = $1
    AND revoked_at IS NULL;

-- name: GetUserFromValidRefreshToken :one
SELECT
    user_id
FROM
    refresh_tokens
WHERE
    token = $1
    AND expires_at > NOW()
    AND revoked_at IS NULL
ORDER BY
    created_at desc;