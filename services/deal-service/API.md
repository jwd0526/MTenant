# CRM Deal Service API Documentation

A comprehensive REST API for managing deals in a multi-tenant CRM platform with automatic tenant isolation, JWT authentication, and real-time analytics.

## 🏗️ API Overview

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
All endpoints require JWT authentication via the `Authorization` header:
```
Authorization: Bearer <jwt-token>
```

### Multi-Tenant Architecture
- **Automatic Tenant Isolation**: All requests are automatically scoped to the authenticated user's tenant
- **Schema-per-Tenant**: Each tenant gets their own PostgreSQL schema for complete data isolation
- **No Tenant Headers Required**: Tenant context is extracted from JWT claims

### Content Type
```
Content-Type: application/json
```

## 📊 Core Resources

### Deal Stages
The API supports the following deal stages:
- `Lead` - Initial prospect
- `Qualified` - Qualified opportunity  
- `Proposal` - Proposal sent
- `Negotiation` - In negotiation
- `Closed Won` - Successfully closed
- `Closed Lost` - Lost opportunity

## 🔧 Endpoints

### 1. Create Deal

Creates a new deal in the authenticated user's tenant.

```http
POST /api/v1/deals
```

**Request Body:**
```json
{
  "title": "Enterprise Software License",
  "value": 250000.00,
  "probability": 75.0,
  "stage": "Qualified",
  "primary_contact_id": 123,
  "company_id": 456,
  "expected_close_date": "2024-03-15T00:00:00Z",
  "deal_source": "Inbound Lead",
  "description": "Multi-year enterprise software license agreement",
  "notes": "Customer interested in our premium package"
}
```

**Request Schema:**
| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `title` | string | ✅ | 1-200 chars | Deal title/name |
| `value` | number | | ≥ 0 | Deal value in currency units |
| `probability` | number | | 0-100 | Win probability percentage |
| `stage` | string | ✅ | Valid stage | Current deal stage |
| `primary_contact_id` | integer | | | Primary contact ID |
| `company_id` | integer | | | Associated company ID |
| `expected_close_date` | datetime | | ISO 8601 | Expected close date |
| `deal_source` | string | | ≤ 100 chars | Lead source |
| `description` | string | | ≤ 1000 chars | Deal description |
| `notes` | string | | ≤ 2000 chars | Internal notes |

**Success Response (201):**
```json
{
  "id": 789,
  "title": "Enterprise Software License",
  "value": 250000.00,
  "probability": 75.0,
  "stage": "Qualified",
  "primary_contact_id": 123,
  "company_id": 456,
  "owner_id": 100,
  "expected_close_date": "2024-03-15T00:00:00Z",
  "actual_close_date": null,
  "deal_source": "Inbound Lead",
  "description": "Multi-year enterprise software license agreement",
  "notes": "Customer interested in our premium package",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "created_by": 100,
  "updated_by": null,
  "primary_contact_name": "John Smith",
  "company_name": "Acme Corp",
  "owner_name": "Jane Doe",
  "deal_age_days": 0,
  "days_until_close": 59,
  "weighted_value": 187500.00
}
```

**Error Responses:**
- `400` - Invalid request data or validation errors
- `401` - Missing or invalid JWT token
- `500` - Internal server error

### 2. Get Deal by ID

Retrieves a single deal with related data and calculated fields.

```http
GET /api/v1/deals/{id}
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | ✅ | Deal ID |

**Success Response (200):**
```json
{
  "id": 789,
  "title": "Enterprise Software License",
  "value": 250000.00,
  "probability": 75.0,
  "stage": "Qualified",
  "primary_contact_id": 123,
  "company_id": 456,
  "owner_id": 100,
  "expected_close_date": "2024-03-15T00:00:00Z",
  "actual_close_date": null,
  "deal_source": "Inbound Lead",
  "description": "Multi-year enterprise software license agreement",
  "notes": "Customer interested in our premium package",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "created_by": 100,
  "updated_by": 100,
  "primary_contact_name": "John Smith",
  "company_name": "Acme Corp",
  "owner_name": "Jane Doe",
  "deal_age_days": 15,
  "days_until_close": 44,
  "weighted_value": 187500.00
}
```

**Calculated Fields:**
- `deal_age_days`: Days since deal creation
- `days_until_close`: Days until expected close date (null if no expected date)
- `weighted_value`: `value * probability / 100`

**Error Responses:**
- `400` - Invalid deal ID format
- `401` - Missing or invalid JWT token
- `404` - Deal not found
- `500` - Internal server error

### 3. Update Deal

Updates an existing deal with partial data (PATCH-style semantics).

```http
PUT /api/v1/deals/{id}
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | ✅ | Deal ID |

**Request Body (all fields optional):**
```json
{
  "title": "Updated Enterprise Software License",
  "value": 300000.00,
  "probability": 85.0,
  "stage": "Proposal",
  "expected_close_date": "2024-03-20T00:00:00Z",
  "notes": "Updated notes with new information"
}
```

**Request Schema:**
All fields from the create request are available as optional fields for partial updates.

**Success Response (200):**
```json
{
  "id": 789,
  "title": "Updated Enterprise Software License",
  "value": 300000.00,
  "probability": 85.0,
  "stage": "Proposal",
  "primary_contact_id": 123,
  "company_id": 456,
  "owner_id": 100,
  "expected_close_date": "2024-03-20T00:00:00Z",
  "actual_close_date": null,
  "deal_source": "Inbound Lead",
  "description": "Multi-year enterprise software license agreement",
  "notes": "Updated notes with new information",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-30T14:20:00Z",
  "created_by": 100,
  "updated_by": 100,
  "primary_contact_name": "John Smith",
  "company_name": "Acme Corp",
  "owner_name": "Jane Doe",
  "deal_age_days": 15,
  "days_until_close": 49,
  "weighted_value": 255000.00
}
```

**Error Responses:**
- `400` - Invalid deal ID or request data
- `401` - Missing or invalid JWT token
- `404` - Deal not found
- `500` - Internal server error

### 4. List Deals

Retrieves a paginated list of deals with filtering and sorting options.

```http
GET /api/v1/deals
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (≥ 1) |
| `limit` | integer | 10 | Items per page (1-100) |
| `stage` | string | | Filter by stage |
| `owner_id` | integer | | Filter by owner |
| `company_id` | integer | | Filter by company |
| `expected_close_from` | date | | Filter by expected close date (≥) |
| `expected_close_to` | date | | Filter by expected close date (≤) |

**Example Request:**
```http
GET /api/v1/deals?page=1&limit=20&stage=Qualified&owner_id=100&expected_close_from=2024-01-01&expected_close_to=2024-03-31
```

**Success Response (200):**
```json
{
  "deals": [
    {
      "id": 789,
      "title": "Enterprise Software License",
      "value": 250000.00,
      "probability": 75.0,
      "stage": "Qualified",
      "primary_contact_id": 123,
      "company_id": 456,
      "owner_id": 100,
      "expected_close_date": "2024-03-15T00:00:00Z",
      "actual_close_date": null,
      "deal_source": "Inbound Lead",
      "description": "Multi-year enterprise software license agreement",
      "notes": "Customer interested in our premium package",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "created_by": 100,
      "updated_by": null,
      "primary_contact_name": "John Smith",
      "company_name": "Acme Corp",
      "owner_name": "Jane Doe",
      "deal_age_days": 15,
      "days_until_close": 44,
      "weighted_value": 187500.00
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_count": 45,
    "total_pages": 3
  }
}
```

**Error Responses:**
- `400` - Invalid query parameters
- `401` - Missing or invalid JWT token
- `500` - Internal server error

### 5. Get Pipeline View

Retrieves deals grouped by pipeline stage with analytics.

```http
GET /api/v1/deals/pipeline
```

**Success Response (200):**
```json
{
  "stages": [
    {
      "stage": "Lead",
      "deal_count": 15,
      "total_value": 500000.00,
      "weighted_value": 125000.00,
      "deals": [
        {
          "id": 101,
          "title": "Small Business Package",
          "value": 15000.00,
          "probability": 25.0,
          "stage": "Lead",
          "primary_contact_id": 201,
          "company_id": 301,
          "owner_id": 100,
          "expected_close_date": "2024-02-28T00:00:00Z",
          "actual_close_date": null,
          "deal_source": "Cold Outreach",
          "description": "Basic package for small business",
          "notes": "Initial contact made",
          "created_at": "2024-01-20T09:15:00Z",
          "updated_at": "2024-01-20T09:15:00Z",
          "created_by": 100,
          "updated_by": null,
          "primary_contact_name": "Alice Johnson",
          "company_name": "Small Corp",
          "owner_name": "Jane Doe",
          "deal_age_days": 10,
          "days_until_close": 28,
          "weighted_value": 3750.00
        }
      ]
    },
    {
      "stage": "Qualified",
      "deal_count": 8,
      "total_value": 1200000.00,
      "weighted_value": 720000.00,
      "deals": [...]
    }
  ],
  "totals": {
    "total_deals": 45,
    "total_value": 3500000.00,
    "total_weighted_value": 1890000.00
  }
}
```

**Response Fields:**
- `stages`: Array of pipeline stages with deals
- `totals`: Aggregate totals across all stages
- Each stage includes deal count, total value, weighted value, and deals array
- Deals are sorted by creation date (newest first)

**Error Responses:**
- `401` - Missing or invalid JWT token
- `500` - Internal server error

### 6. Get Deals by Owner

Retrieves deals assigned to a specific owner.

```http
GET /api/v1/deals/owner/{owner_id}
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `owner_id` | integer | ✅ | Owner/user ID |

**Success Response (200):**
```json
[
  {
    "id": 789,
    "title": "Enterprise Software License",
    "value": 250000.00,
    "probability": 75.0,
    "stage": "Qualified",
    "primary_contact_id": 123,
    "company_id": 456,
    "owner_id": 100,
    "expected_close_date": "2024-03-15T00:00:00Z",
    "actual_close_date": null,
    "deal_source": "Inbound Lead",
    "description": "Multi-year enterprise software license agreement",
    "notes": "Customer interested in our premium package",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "created_by": 100,
    "updated_by": null,
    "primary_contact_name": "John Smith",
    "company_name": "Acme Corp",
    "owner_name": "Jane Doe",
    "deal_age_days": 15,
    "days_until_close": 44,
    "weighted_value": 187500.00
  }
]
```

**Notes:**
- Only returns open deals (no `actual_close_date`)
- Results sorted by expected close date (ascending)
- Empty array if no deals found

**Error Responses:**
- `400` - Invalid owner ID format
- `401` - Missing or invalid JWT token
- `500` - Internal server error

### 7. Close Deal

Closes a deal with final stage and close date.

```http
PUT /api/v1/deals/{id}/close
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | ✅ | Deal ID |

**Request Body:**
```json
{
  "stage": "Closed Won",
  "actual_close_date": "2024-01-30T15:30:00Z"
}
```

**Request Schema:**
| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `stage` | string | ✅ | "Closed Won" or "Closed Lost" | Final stage |
| `actual_close_date` | datetime | | ISO 8601, defaults to now | Actual close date |

**Success Response (200):**
```json
{
  "id": 789,
  "title": "Enterprise Software License",
  "value": 250000.00,
  "probability": 75.0,
  "stage": "Closed Won",
  "primary_contact_id": 123,
  "company_id": 456,
  "owner_id": 100,
  "expected_close_date": "2024-03-15T00:00:00Z",
  "actual_close_date": "2024-01-30T15:30:00Z",
  "deal_source": "Inbound Lead",
  "description": "Multi-year enterprise software license agreement",
  "notes": "Customer interested in our premium package",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-30T15:30:00Z",
  "created_by": 100,
  "updated_by": 100,
  "primary_contact_name": "John Smith",
  "company_name": "Acme Corp",
  "owner_name": "Jane Doe",
  "deal_age_days": 15,
  "days_until_close": null,
  "weighted_value": 187500.00
}
```

**Error Responses:**
- `400` - Invalid deal ID or request data
- `401` - Missing or invalid JWT token  
- `404` - Deal not found
- `500` - Internal server error

### 8. Delete Deal

Permanently deletes an existing deal with automatic tenant isolation.

```http
DELETE /api/v1/deals/{id}
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | ✅ | Deal ID |

**Success Response (204):**
```
204 No Content
```

**Notes:**
- Permanently removes the deal from the database
- This action cannot be undone
- All deal data including relationships will be deleted
- Returns no response body on successful deletion

**Error Responses:**
- `400` - Invalid deal ID format
- `401` - Missing or invalid JWT token
- `404` - Deal not found
- `500` - Internal server error

## 🚨 Error Handling

### Error Response Format

All API errors return a consistent JSON structure:

```json
{
  "error": {
    "type": "validation_error",
    "message": "Title is required and must be between 1 and 200 characters",
    "code": "DEAL_001"
  }
}
```

### HTTP Status Codes

| Code | Description | When Used |
|------|-------------|-----------|
| `200` | OK | Successful GET, PUT requests |
| `201` | Created | Successful POST requests |
| `204` | No Content | Successful DELETE requests |
| `400` | Bad Request | Invalid input, validation errors |
| `401` | Unauthorized | Missing/invalid JWT token |
| `404` | Not Found | Resource not found |
| `500` | Internal Server Error | Server/database errors |

### Common Error Types

| Error Type | Description | HTTP Code |
|------------|-------------|-----------|
| `validation_error` | Request validation failed | 400 |
| `authentication_error` | JWT token issues | 401 |
| `not_found_error` | Resource not found | 404 |
| `database_error` | Database operation failed | 500 |
| `internal_error` | Unexpected server error | 500 |

### Validation Errors

Field validation errors provide specific details:

```json
{
  "error": {
    "type": "validation_error",
    "message": "Validation failed for field 'probability': value must be between 0 and 100",
    "code": "VALIDATION_001"
  }
}
```

## 🔐 Authentication & Security

### JWT Token Requirements

All API endpoints require a valid JWT token in the Authorization header:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Required JWT Claims

The JWT token must contain the following claims:

```json
{
  "user_id": "100",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "permissions": ["deal:read", "deal:write", "deal:delete"],
  "exp": 1642723200,
  "iat": 1642636800
}
```

### Multi-Tenant Isolation

- **Automatic**: Tenant context is automatically extracted from JWT claims
- **Secure**: Complete schema-level isolation prevents cross-tenant data access
- **Transparent**: No tenant headers or parameters required in API calls

### Permission System

| Permission | Description | Required For |
|------------|-------------|--------------|
| `deal:read` | Read deal data | GET endpoints |
| `deal:write` | Create/update deals | POST, PUT endpoints |
| `deal:delete` | Delete deals | DELETE endpoints (future) |

## 📊 Data Types & Formats

### Date/Time Format
- **ISO 8601**: `2024-01-30T15:30:00Z`
- **Timezone**: All timestamps in UTC
- **Date Only**: `2024-01-30` (for date filters)

### Numeric Values
- **Currency**: Decimal numbers with up to 2 decimal places
- **Percentages**: 0-100 (probability values)
- **IDs**: Positive integers

### String Lengths
| Field | Max Length | Notes |
|-------|------------|-------|
| `title` | 200 | Required |
| `deal_source` | 100 | Optional |
| `description` | 1000 | Optional |
| `notes` | 2000 | Optional |

## 🔄 Workflow Examples

### Creating and Managing a Deal

1. **Create Deal**:
```http
POST /api/v1/deals
{
  "title": "Q1 Enterprise Deal",
  "value": 500000,
  "probability": 60,
  "stage": "Lead",
  "company_id": 123,
  "expected_close_date": "2024-03-31T23:59:59Z"
}
```

2. **Update as it Progresses**:
```http
PUT /api/v1/deals/789
{
  "stage": "Qualified",
  "probability": 75,
  "notes": "Demo completed successfully"
}
```

3. **Close the Deal**:
```http
PUT /api/v1/deals/789/close
{
  "stage": "Closed Won",
  "actual_close_date": "2024-03-25T14:30:00Z"
}
```

### Sales Pipeline Management

1. **View Pipeline**:
```http
GET /api/v1/deals/pipeline
```

2. **Filter by Stage**:
```http
GET /api/v1/deals?stage=Negotiation&limit=50
```

3. **Track Owner Performance**:
```http
GET /api/v1/deals/owner/100
```

## 🚀 Performance Considerations

### Pagination
- **Default**: 10 items per page
- **Maximum**: 100 items per page  
- **Recommendation**: Use 20-50 items for optimal performance

### Filtering
- **Database-Level**: All filtering happens at the database level
- **Indexed Fields**: `stage`, `owner_id`, `company_id`, `expected_close_date`
- **Performance**: Filtered queries typically execute in <50ms

### Caching
- **Response Caching**: Not implemented (real-time data)
- **Connection Pooling**: Efficient PostgreSQL connection reuse
- **Query Optimization**: All queries use prepared statements

## 🛠️ Development & Testing

### Local Development Setup

1. **Start the service**:
```bash
export DATABASE_URL="postgres://user:pass@localhost:5433/crm-platform"
export SHARED_JWT_SECRET="your-secret-key"
go run cmd/server/main.go
```

2. **Get JWT Token** (for testing):
```bash
# Use your authentication service or create a test token
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

3. **Test API calls**:
```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/v1/deals
```

### Postman Collection

Import the provided Postman collection for complete API testing:
- Environment variables for base URL and JWT token
- Complete request examples for all endpoints
- Test assertions for response validation

### API Testing

```bash
# Health check
curl http://localhost:8080/health

# Create deal
curl -X POST \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Deal", "value": 1000, "stage": "Lead"}' \
  http://localhost:8080/api/v1/deals

# List deals
curl -H "Authorization: Bearer $JWT_TOKEN" \
  "http://localhost:8080/api/v1/deals?page=1&limit=10"
```

## 📝 Changelog & Versioning

### API Version: v0
- **Current Version**: 0.1.1
- **Stability**: Stable
- **Backward Compatibility**: Guaranteed

### Recent Changes
- **v1.0.0**: Initial API release with full CRUD operations
- **Multi-tenant**: Schema-per-tenant isolation
- **Authentication**: JWT-based security
- **Analytics**: Pipeline view and calculated fields

---

This API provides comprehensive deal management capabilities with enterprise-grade security, performance, and multi-tenant isolation. For support or questions, please refer to the repository documentation or create an issue.