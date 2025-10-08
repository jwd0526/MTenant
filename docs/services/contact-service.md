# Contact Service (Planned)

**Last Updated:** 2025-10-08\
*SQLC setup complete, handlers and business logic pending*

Customer and company data management service.

## Overview

The Contact Service manages individual contacts and company records with support for hierarchical relationships, custom fields, advanced search capabilities, and data import/export functionality.

## Current Implementation Status

**Status**: Basic SQLC setup completed, placeholder main.go implementation
- ✅ SQLC configuration (`sqlc.yaml`) 
- ✅ Database schema (`db/schema/`)
- ✅ SQL queries (`db/queries/`)
- ✅ Generated code (`internal/db/`)
- ❌ HTTP handlers and business logic (planned)
- ❌ Search and filtering logic (planned)
- ❌ Import/export functionality (planned)

## Database Schema

### Tenant-Specific Tables

**`contacts`** - Individual contact records
```sql
CREATE TABLE contacts (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    title VARCHAR(100),
    company_id INTEGER REFERENCES companies(id),
    status VARCHAR(50) DEFAULT 'active',
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    updated_by INTEGER
);
```

**`companies`** - Business entities with hierarchical support
```sql
CREATE TABLE companies (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(255) NOT NULL,
    website VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    industry VARCHAR(100),
    size VARCHAR(50),
    parent_company_id INTEGER REFERENCES companies(id),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    updated_by INTEGER
);
```

## SQLC Queries

The service includes comprehensive SQLC queries for contact and company management:

- **Contact Management**: `CreateContact`, `GetContactByID`, `UpdateContact`, `DeleteContact`, `ListContacts`
- **Contact Search**: `SearchContactsByName`, `SearchContactsByEmail`, `GetContactsByCompany`
- **Company Management**: `CreateCompany`, `GetCompanyByID`, `UpdateCompany`, `DeleteCompany`, `ListCompanies`
- **Company Relationships**: `GetSubsidiaries`, `GetParentCompany`, `GetCompanyHierarchy`
- **Association Management**: `AssociateContactWithCompany`, `GetContactsWithCompanyInfo`

## Planned API Endpoints

**Current Status**: Endpoints not implemented - service has placeholder main.go

### Contact Management
```
POST   /api/contacts               # Create new contact
GET    /api/contacts/:id           # Get contact details
PUT    /api/contacts/:id           # Update contact
DELETE /api/contacts/:id           # Soft delete contact
GET    /api/contacts               # List/search contacts with pagination
```

### Contact Search & Filtering
```
GET    /api/contacts/search        # Advanced search with filters
GET    /api/contacts/company/:id   # Get contacts by company
GET    /api/contacts/export        # Export contacts (CSV/Excel)
POST   /api/contacts/import        # Bulk import contacts
```

### Company Management
```
POST   /api/companies              # Create new company
GET    /api/companies/:id          # Get company details
PUT    /api/companies/:id          # Update company
DELETE /api/companies/:id          # Soft delete company
GET    /api/companies              # List companies with pagination
```

### Company Relationships
```
GET    /api/companies/:id/contacts     # Get company contacts
GET    /api/companies/:id/subsidiaries # Get subsidiary companies
GET    /api/companies/:id/hierarchy    # Get company hierarchy
POST   /api/companies/:id/subsidiaries # Add subsidiary relationship
```

## Planned Features

### Contact Management
- Full CRUD operations with validation
- Custom field support (JSONB)
- Contact-company association management
- Contact status tracking (active, inactive, archived)
- Contact notes and interaction history

### Company Management
- Hierarchical company relationships (parent/subsidiary)
- Company size and industry categorization
- Complete address management
- Company-level custom fields
- Company status and lifecycle management

### Search and Filtering
- Full-text search across contact fields
- Advanced filtering by company, industry, status
- Date range filtering (created, updated)
- Custom field filtering
- Pagination and sorting options

### Data Import/Export
- CSV and Excel import/export
- Field mapping for imports
- Bulk validation and error reporting
- Export with custom field selection
- Template download for imports

## Service Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://admin:admin@localhost:5433/crm-platform

# File Storage (for imports/exports)
STORAGE_TYPE=local # or s3, gcs
UPLOAD_MAX_SIZE=10MB
TEMP_DIR=/tmp/uploads

# Search Configuration
SEARCH_INDEX_ENABLED=true
SEARCH_RESULTS_LIMIT=100

# Application
PORT=8082
LOG_LEVEL=info
```

## Data Validation

### Contact Validation
- Email format validation
- Phone number format validation
- Required field enforcement
- Custom field type validation
- Duplicate email checking

### Company Validation
- Company name uniqueness within tenant
- Website URL format validation
- Parent company relationship validation
- Industry and size enumeration validation
- Address format validation

## Search Capabilities

### Full-Text Search
- Search across name, email, company, title
- Fuzzy matching for typos
- Weighted search results
- Search highlighting

### Advanced Filtering
```json
{
  "filters": {
    "company_id": 123,
    "status": ["active", "inactive"],
    "industry": ["technology", "healthcare"],
    "created_date": {
      "from": "2024-01-01",
      "to": "2024-12-31"
    },
    "custom_fields": {
      "lead_source": "website"
    }
  },
  "sort": {
    "field": "last_name",
    "direction": "asc"
  },
  "pagination": {
    "page": 1,
    "limit": 50
  }
}
```

## Import/Export Features

### Supported Formats
- CSV (comma-separated values)
- Excel (.xlsx)
- JSON (for API integrations)

### Import Features
- Field mapping interface
- Data validation and preview
- Error reporting and correction
- Batch processing for large files
- Duplicate detection and handling

### Export Features
- Custom field selection
- Filter-based exports
- Template generation
- Scheduled exports (planned)

## Inter-Service Communication

### Outbound Calls (Planned)
- **Auth Service**: User authentication and context
- **Deal Service**: Contact-deal associations
- **Communication Service**: Contact activity logging
- **File Service**: Import/export file handling

### Inbound Calls (Planned)
- **Deal Service**: Contact information for deals
- **Communication Service**: Contact details for activities
- **Frontend**: Contact and company management

## Performance Considerations

### Database Optimization
- Indexes on frequently searched fields (email, name, company_id)
- JSONB indexes for custom field searches
- Partial indexes for status filtering
- Connection pooling for concurrent requests

### Search Performance
- Database full-text search capabilities
- Indexed custom fields for fast filtering
- Pagination to limit result sets
- Caching for common search queries

### Import/Export Performance
- Streaming for large file processing
- Background job processing for imports
- Chunked processing to prevent timeouts
- Progress tracking for user feedback

## Security Considerations

### Data Protection
- Tenant isolation for all contact data
- Field-level encryption for sensitive data
- Audit logging for all modifications
- GDPR compliance for data handling

### Access Control
- Role-based contact access
- Company-level permissions
- Contact ownership and sharing
- Export permission controls

## Testing Strategy

### Current Tests
- Basic SQLC generated code tests
- Database connection tests
- Utility function tests

### Planned Tests
- Contact CRUD operation tests
- Company relationship tests
- Search and filtering functionality tests
- Import/export workflow tests
- Performance tests with large datasets

## Development Status

**Current Directory Structure:**
```
services/contact-service/
├── cmd/server/
│   ├── main.go                 # Placeholder implementation
│   └── main_test.go           # Basic tests
├── internal/
│   ├── db/                    # Generated SQLC code
│   ├── benchmark_test.go      # Performance tests
│   └── utils_test.go          # Utility tests
├── db/
│   ├── queries/               # SQL query files
│   └── schema/               # Database schema
├── Dockerfile                 # Container definition
├── go.mod                    # Go dependencies
└── sqlc.yaml                 # SQLC configuration
```

## Next Implementation Steps

1. **CRUD Operations**: Implement basic contact and company operations
2. **Search Implementation**: Add search and filtering capabilities
3. **HTTP Handlers**: Create REST API endpoints
4. **Validation Layer**: Add comprehensive data validation
5. **Import/Export**: Build file processing capabilities
6. **Relationship Management**: Implement company hierarchy features
7. **Performance Optimization**: Add indexing and caching
8. **Integration Testing**: Test with deal and communication services

## Related Documentation

- [Service Architecture](../architecture/services.md) - Overall microservices design
- [Database Design](../architecture/database.md) - Multi-tenant data architecture
- [SQLC Implementation](../architecture/sqlc.md) - Database query patterns
- [Contact Service Queries](../database/queries/contact-service.md) - SQL query documentation