-- name: AddDealContact :one
INSERT INTO deal_contacts (deal_id, contact_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetDealContacts :many
SELECT c.id, c.first_name, c.last_name, c.email, dc.role
FROM contacts c
JOIN deal_contacts dc ON c.id = dc.contact_id
WHERE dc.deal_id = $1;

-- name: RemoveDealContact :exec
DELETE FROM deal_contacts 
WHERE deal_id = $1 AND contact_id = $2;

-- name: GetContactDeals :many
SELECT d.id, d.title, d.value, d.stage, dc.role
FROM deals d
JOIN deal_contacts dc ON d.id = dc.deal_id
WHERE dc.contact_id = $1;