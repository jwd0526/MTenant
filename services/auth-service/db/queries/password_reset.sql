-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, token, expires_at, created_at;

-- name: GetPasswordResetToken :one
SELECT prt.id, prt.user_id, prt.token, prt.expires_at, prt.used_at,
       u.email, u.first_name, u.last_name, u.active
FROM password_reset_tokens prt
JOIN users u ON prt.user_id = u.id
WHERE prt.token = $1 
  AND prt.used_at IS NULL 
  AND prt.expires_at > CURRENT_TIMESTAMP
  AND u.active = true 
  AND u.deleted_at IS NULL;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = CURRENT_TIMESTAMP
WHERE token = $1;

-- name: CleanupExpiredTokens :exec
DELETE FROM password_reset_tokens
WHERE expires_at < CURRENT_TIMESTAMP OR used_at IS NOT NULL;

-- name: GetUserPasswordResetTokens :many
SELECT id, token, expires_at, used_at, created_at
FROM password_reset_tokens
WHERE user_id = $1
ORDER BY created_at DESC;