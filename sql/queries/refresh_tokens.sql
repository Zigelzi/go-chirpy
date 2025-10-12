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