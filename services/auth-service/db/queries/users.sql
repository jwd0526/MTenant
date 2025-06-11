-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, role, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, first_name, last_name, role, active, email_verified, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, first_name, last_name, role, permissions, 
       active, email_verified, last_login, created_at, updated_at
FROM users
WHERE email = $1 AND active = true AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, role, permissions, 
       active, email_verified, last_login, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserForAuth :one
SELECT id, password_hash, role, active, email_verified
FROM users 
WHERE email = $1 AND active = true AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users 
SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = true, updated_at = CURRENT_TIMESTAMP, updated_by = $2
WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE users
SET role = $2, updated_at = CURRENT_TIMESTAMP, updated_by = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserPermissions :exec
UPDATE users
SET permissions = $2, updated_at = CURRENT_TIMESTAMP, updated_by = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = CURRENT_TIMESTAMP, updated_by = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeactivateUser :exec
UPDATE users
SET active = false, updated_at = CURRENT_TIMESTAMP, updated_by = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP, updated_by = $2
WHERE id = $1;

-- name: ListActiveUsers :many
SELECT id, email, first_name, last_name, role, active, email_verified, 
       last_login, created_at
FROM users
WHERE active = true AND deleted_at IS NULL
ORDER BY first_name, last_name;

-- name: ListUsersByRole :many
SELECT id, email, first_name, last_name, role, active, email_verified,
       last_login, created_at
FROM users
WHERE role = $1 AND active = true AND deleted_at IS NULL
ORDER BY first_name, last_name;

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE email = $1 AND deleted_at IS NULL
);