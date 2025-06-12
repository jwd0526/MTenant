-- name: CreateInvitation :one
INSERT INTO invitations (tenant_id, email, role, token, expires_at, invited_by, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, token, expires_at, created_at;

-- name: GetInvitationByToken :one
SELECT i.id, i.tenant_id, i.email, i.role, i.token, i.expires_at, i.accepted_at, i.metadata,
       t.name as tenant_name, t.subdomain as tenant_subdomain, t.schema_name as tenant_schema
FROM invitations i
JOIN tenants t ON i.tenant_id = t.id
WHERE i.token = $1 
  AND i.accepted_at IS NULL 
  AND i.expires_at > CURRENT_TIMESTAMP;

-- name: AcceptInvitation :exec
UPDATE invitations
SET accepted_at = CURRENT_TIMESTAMP
WHERE token = $1 AND accepted_at IS NULL;

-- name: ListTenantInvitations :many
SELECT id, email, role, token, expires_at, accepted_at, created_at
FROM invitations
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListPendingInvitations :many
SELECT i.id, i.email, i.role, i.expires_at, i.created_at,
       t.name as tenant_name, t.subdomain as tenant_subdomain
FROM invitations i
JOIN tenants t ON i.tenant_id = t.id
WHERE i.accepted_at IS NULL 
  AND i.expires_at > CURRENT_TIMESTAMP
ORDER BY i.created_at DESC;

-- name: CheckPendingInvitation :one
SELECT EXISTS(
    SELECT 1 FROM invitations 
    WHERE tenant_id = $1 
      AND email = $2 
      AND accepted_at IS NULL 
      AND expires_at > CURRENT_TIMESTAMP
);

-- name: CleanupExpiredInvitations :exec
DELETE FROM invitations
WHERE expires_at < CURRENT_TIMESTAMP;

-- name: RevokeInvitation :exec
DELETE FROM invitations
WHERE id = $1 AND tenant_id = $2 AND accepted_at IS NULL;

-- name: GetInvitationsByEmail :many
SELECT i.id, i.tenant_id, i.role, i.token, i.expires_at, i.accepted_at,
       t.name as tenant_name, t.subdomain as tenant_subdomain
FROM invitations i
JOIN tenants t ON i.tenant_id = t.id
WHERE i.email = $1
ORDER BY i.created_at DESC;