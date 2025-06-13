-- name: CreateContact :one
INSERT INTO contacts (
    first_name, last_name, email, phone, company_id, custom_fields, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetContactByID :one
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.id = $1 AND c.deleted_at IS NULL;

-- name: GetContactByEmail :one
SELECT id, first_name, last_name, company_id FROM contacts 
WHERE email = $1 AND deleted_at IS NULL;

-- name: ListContacts :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL
ORDER BY c.last_name, c.first_name
LIMIT $1 OFFSET $2;

-- name: UpdateContact :one
UPDATE contacts 
SET first_name = $2, last_name = $3, email = $4, phone = $5, 
    company_id = $6, custom_fields = $7, updated_by = $8, 
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteContact :exec
UPDATE contacts 
SET deleted_at = CURRENT_TIMESTAMP, updated_by = $2 
WHERE id = $1;

-- name: SearchContactsFullText :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL 
  AND to_tsvector('english', c.first_name || ' ' || c.last_name || ' ' || COALESCE(c.email, '')) 
      @@ to_tsquery('english', $1)
ORDER BY c.last_name, c.first_name
LIMIT $2 OFFSET $3;

-- name: ListContactsByCompany :many
SELECT * FROM contacts 
WHERE company_id = $1 AND deleted_at IS NULL
ORDER BY last_name, first_name;

-- name: FilterContacts :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL
  AND ($1::int IS NULL OR c.company_id = $1)
  AND ($2::text IS NULL OR c.custom_fields->>'status' = $2)
ORDER BY c.last_name, c.first_name
LIMIT $3 OFFSET $4;

-- name: CountContacts :one
SELECT COUNT(*) FROM contacts WHERE deleted_at IS NULL;

-- name: CountContactsByCompany :one
SELECT COUNT(*) FROM contacts 
WHERE company_id = $1 AND deleted_at IS NULL;

-- name: UpdateContactCustomFields :one
UPDATE contacts 
SET custom_fields = $2, updated_by = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SearchContactsByCustomField :many
SELECT * FROM contacts 
WHERE custom_fields->>$1 = $2 AND deleted_at IS NULL
ORDER BY last_name, first_name;

-- name: GetContactsByDomain :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.email LIKE '%@' || $1 AND c.deleted_at IS NULL
ORDER BY c.last_name, c.first_name;