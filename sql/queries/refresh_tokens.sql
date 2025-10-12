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

-- name: RevokeRefreshTokenByUserId :execrows
UPDATE refresh_tokens
SET
    updated_at = NOW(),
    revoked_at = NOW()
WHERE
    user_id = $1
    AND revoked_at IS NULL
RETURNING
    *;