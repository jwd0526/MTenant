# Contact Service SQL Queries

**Last Updated:** 2025-10-08\
*SQLC query documentation for contact-service*

Comprehensive documentation of SQL queries used by the Contact Service for managing contacts and companies.

## Overview

The Contact Service handles customer and company data management within tenant schemas. It supports full CRUD operations, advanced search, and relationship management.

## Contact Management Queries

### Contact Creation

**Query: `CreateContact`**
```sql
-- name: CreateContact :one
INSERT INTO contacts (
    first_name, last_name, email, phone, company_id, custom_fields, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;
```

**Purpose:** Create a new contact with optional company association and custom fields.

**Parameters:**
- `$1` - First name (VARCHAR 100, required)
- `$2` - Last name (VARCHAR 100, required)
- `$3` - Email address (VARCHAR 254, nullable)
- `$4` - Phone number (VARCHAR 20, nullable)
- `$5` - Company ID (INTEGER, nullable FK)
- `$6` - Custom fields (JSONB)
- `$7` - Created by user ID (INTEGER, required)

**Generated Go Method:**
```go
type CreateContactParams struct {
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Email        *string `json:"email"`
    Phone        *string `json:"phone"`
    CompanyID    *int32 `json:"company_id"`
    CustomFields []byte `json:"custom_fields"`
    CreatedBy    int32 `json:"created_by"`
}

func (q *Queries) CreateContact(ctx context.Context, arg CreateContactParams) (Contact, error)
```

**Usage Example:**
```go
customFields := map[string]interface{}{
    "lead_source": "website",
    "priority": "high",
    "decision_maker": true,
}
customFieldsJSON, _ := json.Marshal(customFields)

contact, err := queries.CreateContact(ctx, db.CreateContactParams{
    FirstName:    "Sarah",
    LastName:     "Johnson",
    Email:        &email,
    Phone:        &phone,
    CompanyID:    &companyID,
    CustomFields: customFieldsJSON,
    CreatedBy:    userID,
})
```

### Contact Retrieval

**Query: `GetContactByID`**
```sql
-- name: GetContactByID :one
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.id = $1 AND c.deleted_at IS NULL;
```

**Purpose:** Get complete contact details with company information.

**Features:**
- Includes company name via LEFT JOIN
- Respects soft delete for both contacts and companies
- Single query for complete contact view

**Query: `GetContactByEmail`**
```sql
-- name: GetContactByEmail :one
SELECT id, first_name, last_name, company_id FROM contacts 
WHERE email = $1 AND deleted_at IS NULL;
```

**Purpose:** Find contact by email for deduplication and validation.

**Usage Example:**
```go
// Check for existing contact before creation
existingContact, err := queries.GetContactByEmail(ctx, "sarah@techflow.com")
if err == nil {
    return errors.New("contact with this email already exists")
}

// Get full contact details
contact, err := queries.GetContactByID(ctx, contactID)
if err != nil {
    return c.JSON(404, gin.H{"error": "Contact not found"})
}
```

### Contact Updates

**Query: `UpdateContact`**
```sql
-- name: UpdateContact :one
UPDATE contacts 
SET first_name = $2, last_name = $3, email = $4, phone = $5, 
    company_id = $6, custom_fields = $7, updated_by = $8, 
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
```

**Purpose:** Update contact information with audit tracking.

**Features:**
- Updates all standard fields
- Preserves audit trail with updated_by and updated_at
- Returns complete updated record
- Respects soft delete constraint

**Query: `UpdateContactCustomFields`**
```sql
-- name: UpdateContactCustomFields :one
UPDATE contacts 
SET custom_fields = $2, updated_by = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
```

**Purpose:** Update only custom fields for performance.

**Usage Example:**
```go
// Update standard fields
updatedContact, err := queries.UpdateContact(ctx, db.UpdateContactParams{
    ID:           contactID,
    FirstName:    "Sarah",
    LastName:     "Johnson-Smith",
    Email:        &newEmail,
    Phone:        &newPhone,
    CompanyID:    &newCompanyID,
    CustomFields: customFieldsJSON,
    UpdatedBy:    &userID,
})

// Update only custom fields
newCustomFields := map[string]interface{}{
    "priority": "urgent",
    "last_activity": time.Now(),
}
customFieldsJSON, _ := json.Marshal(newCustomFields)

contact, err := queries.UpdateContactCustomFields(ctx, db.UpdateContactCustomFieldsParams{
    ID:           contactID,
    CustomFields: customFieldsJSON,
    UpdatedBy:    &userID,
})
```

### Contact Deletion

**Query: `SoftDeleteContact`**
```sql
-- name: SoftDeleteContact :exec
UPDATE contacts 
SET deleted_at = CURRENT_TIMESTAMP, updated_by = $2 
WHERE id = $1;
```

**Purpose:** Soft delete contact while preserving data integrity.

**Features:**
- Maintains foreign key relationships
- Preserves audit trail
- Excludes from future queries automatically

**Usage Example:**
```go
err := queries.SoftDeleteContact(ctx, db.SoftDeleteContactParams{
    ID:        contactID,
    UpdatedBy: &userID,
})
if err != nil {
    return c.JSON(500, gin.H{"error": "Failed to delete contact"})
}
```

## Contact Search and Filtering

### Basic Listing

**Query: `ListContacts`**
```sql
-- name: ListContacts :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL
ORDER BY c.last_name, c.first_name
LIMIT $1 OFFSET $2;
```

**Purpose:** Paginated list of contacts with company names.

**Features:**
- Includes company information
- Alphabetical sorting
- Pagination support
- Soft delete filtering

### Full-Text Search

**Query: `SearchContactsFullText`**
```sql
-- name: SearchContactsFullText :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL 
  AND to_tsvector('english', c.first_name || ' ' || c.last_name || ' ' || COALESCE(c.email, '')) 
      @@ to_tsquery('english', $1)
ORDER BY c.last_name, c.first_name
LIMIT $2 OFFSET $3;
```

**Purpose:** Advanced text search across name and email fields.

**Features:**
- PostgreSQL full-text search with GIN index
- Searches across first name, last name, and email
- Handles partial matches and multiple terms
- Supports advanced search operators

**Search Examples:**
```sql
-- Search for "john smith"
$1 = 'john & smith'

-- Search for "john OR smith"  
$1 = 'john | smith'

-- Search for phrase
$1 = 'john smith'

-- Search with wildcards
$1 = 'john:*'
```

**Usage Example:**
```go
// Parse search query
searchTerms := strings.Fields(query)
tsQuery := strings.Join(searchTerms, " & ")

contacts, err := queries.SearchContactsFullText(ctx, db.SearchContactsFullTextParams{
    ToTsquery: tsQuery,
    Limit:     20,
    Offset:    0,
})
```

### Advanced Filtering

**Query: `FilterContacts`**
```sql
-- name: FilterContacts :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.deleted_at IS NULL
  AND ($1::int IS NULL OR c.company_id = $1)
  AND ($2::text IS NULL OR c.custom_fields->>'status' = $2)
ORDER BY c.last_name, c.first_name
LIMIT $3 OFFSET $4;
```

**Purpose:** Filter contacts by company and custom field criteria.

**Features:**
- Optional company filter
- Custom field filtering (extensible pattern)
- Null-safe parameter handling
- Maintains pagination

**Usage Example:**
```go
// Filter by company
contacts, err := queries.FilterContacts(ctx, db.FilterContactsParams{
    CompanyID: &companyID,
    Status:    nil,
    Limit:     20,
    Offset:    0,
})

// Filter by custom status
status := "qualified"
contacts, err = queries.FilterContacts(ctx, db.FilterContactsParams{
    CompanyID: nil,
    Status:    &status,
    Limit:     20,
    Offset:    0,
})
```

## Company-Specific Queries

### Company Association

**Query: `ListContactsByCompany`**
```sql
-- name: ListContactsByCompany :many
SELECT * FROM contacts 
WHERE company_id = $1 AND deleted_at IS NULL
ORDER BY last_name, first_name;
```

**Purpose:** Get all contacts associated with a specific company.

**Query: `GetContactsByDomain`**
```sql
-- name: GetContactsByDomain :many
SELECT c.*, comp.name as company_name 
FROM contacts c
LEFT JOIN companies comp ON c.company_id = comp.id AND comp.deleted_at IS NULL
WHERE c.email LIKE '%@' || $1 AND c.deleted_at IS NULL
ORDER BY c.last_name, c.first_name;
```

**Purpose:** Find contacts by email domain for company association.

**Usage Example:**
```go
// Get all contacts at TechFlow
companyContacts, err := queries.ListContactsByCompany(ctx, companyID)

// Find contacts by domain for bulk association
domainContacts, err := queries.GetContactsByDomain(ctx, "techflow.com")
for _, contact := range domainContacts {
    if contact.CompanyID == nil {
        // Associate with company
        queries.UpdateContact(ctx, db.UpdateContactParams{
            ID:        contact.ID,
            CompanyID: &companyID,
            // ... other fields
        })
    }
}
```

## Custom Field Queries

### Custom Field Search

**Query: `SearchContactsByCustomField`**
```sql
-- name: SearchContactsByCustomField :many
SELECT * FROM contacts 
WHERE custom_fields->>$1 = $2 AND deleted_at IS NULL
ORDER BY last_name, first_name;
```

**Purpose:** Find contacts by specific custom field values.

**Usage Example:**
```go
// Find high-priority contacts
priorityContacts, err := queries.SearchContactsByCustomField(ctx, db.SearchContactsByCustomFieldParams{
    Key:   "priority",
    Value: "high",
})

// Find contacts from trade shows
tradeShowContacts, err := queries.SearchContactsByCustomField(ctx, db.SearchContactsByCustomFieldParams{
    Key:   "lead_source", 
    Value: "trade_show",
})
```

## Analytics and Counting

### Contact Statistics

**Query: `CountContacts`**
```sql
-- name: CountContacts :one
SELECT COUNT(*) FROM contacts WHERE deleted_at IS NULL;
```

**Query: `CountContactsByCompany`**
```sql
-- name: CountContactsByCompany :one
SELECT COUNT(*) FROM contacts 
WHERE company_id = $1 AND deleted_at IS NULL;
```

**Purpose:** Generate contact statistics for dashboards and reporting.

**Usage Example:**
```go
// Dashboard statistics
totalContacts, err := queries.CountContacts(ctx)
companyContactCount, err := queries.CountContactsByCompany(ctx, companyID)

stats := DashboardStats{
    TotalContacts:        totalContacts,
    CompanyContactCount:  companyContactCount,
}
```

## Company Management Queries

### Company Creation

**Query: `CreateCompany`** (not shown in provided files but standard pattern)
```sql
-- name: CreateCompany :one
INSERT INTO companies (
    name, domain, industry, street, city, state, country, 
    postal_code, custom_fields, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;
```

### Company Retrieval

**Query: `GetCompanyByID`**
```sql
-- name: GetCompanyByID :one
SELECT * FROM companies 
WHERE id = $1 AND deleted_at IS NULL;
```

**Query: `GetCompanyByDomain`**
```sql
-- name: GetCompanyByDomain :one
SELECT * FROM companies 
WHERE domain = $1 AND deleted_at IS NULL;
```

**Purpose:** Company lookup for contact association and deduplication.

## Performance Optimization

### Index Usage

**Critical Indexes:**
- `idx_contacts_email` - Email lookups and domain searches
- `idx_contacts_company_id` - Company associations
- `idx_contacts_search` - Full-text search (GIN index)

**Query Performance:**
```sql
-- Verify email lookup performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT id FROM contacts WHERE email = 'sarah@techflow.com';

-- Verify full-text search performance  
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM contacts 
WHERE to_tsvector('english', first_name || ' ' || last_name || ' ' || COALESCE(email, '')) 
      @@ to_tsquery('english', 'sarah & johnson');
```

### Pagination Best Practices

**Efficient Pagination:**
```go
const DefaultPageSize = 20
const MaxPageSize = 100

func validatePagination(limit, offset int) (int, int) {
    if limit <= 0 || limit > MaxPageSize {
        limit = DefaultPageSize
    }
    if offset < 0 {
        offset = 0
    }
    return limit, offset
}

// Usage
limit, offset := validatePagination(requestLimit, requestOffset)
contacts, err := queries.ListContacts(ctx, db.ListContactsParams{
    Limit:  int32(limit),
    Offset: int32(offset),
})
```

## Error Handling

### Common Error Patterns

**Duplicate Email:**
```go
_, err := queries.CreateContact(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") &&
       strings.Contains(err.Error(), "email") {
        return c.JSON(409, gin.H{"error": "Contact with this email already exists"})
    }
}
```

**Foreign Key Constraints:**
```go
_, err := queries.CreateContact(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "foreign key constraint") &&
       strings.Contains(err.Error(), "company_id") {
        return c.JSON(400, gin.H{"error": "Invalid company ID"})
    }
}
```

**Contact Not Found:**
```go
contact, err := queries.GetContactByID(ctx, contactID)
if errors.Is(err, pgx.ErrNoRows) {
    return c.JSON(404, gin.H{"error": "Contact not found"})
}
```

## Custom Field Patterns

### Flexible Schema Design

**Common Custom Fields:**
```json
{
    "lead_source": "website|trade_show|referral|cold_call",
    "priority": "low|medium|high|urgent",
    "decision_maker": true,
    "budget_authority": false,
    "contact_preference": "email|phone|linkedin",
    "timezone": "America/New_York",
    "tags": ["enterprise", "warm_lead", "decision_maker"],
    "last_activity": "2024-01-15T10:30:00Z",
    "notes": "Interested in enterprise package"
}
```

**Query Patterns:**
```sql
-- Multiple custom field filters
SELECT * FROM contacts 
WHERE custom_fields->>'priority' = 'high'
  AND custom_fields->>'decision_maker' = 'true'
  AND custom_fields ? 'budget_authority';

-- Array contains
SELECT * FROM contacts 
WHERE custom_fields->'tags' ? 'enterprise';

-- Date range filtering
SELECT * FROM contacts 
WHERE (custom_fields->>'last_activity')::timestamp 
      >= NOW() - INTERVAL '30 days';
```

## Related Documentation

- [Contact Schema](../tenant-template/ContactsSchema.md) - Complete table definitions
- [Company Schema](../tenant-template/CompanySchema.md) - Company table structure
- [SQLC Configuration](../../architecture/sqlc.md) - Code generation setup
- [Contact Service Architecture](../../services/contact-service.md) - Service implementation