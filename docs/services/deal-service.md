# Deal Service

**Last Updated:** 2025-10-08\
*Fully implemented with handlers, business logic, and pipeline management*

Sales pipeline and opportunity management service.

## Overview

The Deal Service manages sales opportunities through configurable pipeline stages, revenue forecasting, deal-contact associations, and comprehensive sales analytics for the MTenant CRM platform.

## Current Implementation Status

**Status**: Fully implemented and operational
- ✅ SQLC configuration and generated code
- ✅ Database schema and queries
- ✅ HTTP handlers (`internal/handlers/`)
- ✅ Business logic layer (`internal/business/`)
- ✅ Request/response models (`internal/models/`)
- ✅ Tenant-aware middleware
- ✅ Pipeline management logic
- ✅ Deal-contact associations
- ✅ Integration tests

## Database Schema

### Tenant-Specific Tables

**`deals`** - Sales opportunities with stage tracking
```sql
CREATE TABLE deals (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    value DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    stage VARCHAR(100) NOT NULL,
    probability INTEGER DEFAULT 0, -- 0-100 percentage
    expected_close_date DATE,
    actual_close_date DATE,
    status VARCHAR(50) DEFAULT 'open', -- open, won, lost
    company_id INTEGER, -- Reference to companies table
    contact_id INTEGER, -- Primary contact for the deal
    source VARCHAR(100), -- lead source
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    updated_by INTEGER,
    closed_by INTEGER,
    lost_reason VARCHAR(255)
);
```

**`deal_contacts`** - Many-to-many contact associations
```sql
CREATE TABLE deal_contacts (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    deal_id INTEGER REFERENCES deals(id) ON DELETE CASCADE,
    contact_id INTEGER NOT NULL, -- Reference to contacts table
    role VARCHAR(100), -- decision_maker, influencer, user, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(deal_id, contact_id)
);
```

**`deal_stages`** - Configurable pipeline stages (planned)
```sql
CREATE TABLE deal_stages (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(100) NOT NULL,
    order_index INTEGER NOT NULL,
    probability INTEGER DEFAULT 0,
    is_closed BOOLEAN DEFAULT false,
    is_won BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_index)
);
```

## SQLC Configuration

The service uses SQLC with decimal support for financial calculations:

```yaml
# sqlc.yaml
overrides:
  - column: "*.value"
    go_type: "github.com/shopspring/decimal.Decimal"
  - column: "*.expected_close_date"
    go_type: "database/sql.NullTime"
  - column: "*.actual_close_date"
    go_type: "database/sql.NullTime"
```

## SQLC Queries

The service includes comprehensive type-safe queries:

- **Deal Management**: `CreateDeal`, `GetDealByID`, `UpdateDeal`, `DeleteDeal`, `ListDeals`
- **Pipeline Operations**: `GetDealsByStage`, `GetPipelineOverview`
- **Deal Analytics**: `GetDealsByDateRange`, `GetWonDealsTotal`
- **Owner Operations**: `GetDealsByOwner`

## API Endpoints

All endpoints are fully implemented and operational.

### Deal Management
```
POST   /api/v1/deals               # Create new deal
GET    /api/v1/deals/:id           # Get deal details
PUT    /api/v1/deals/:id           # Update deal
DELETE /api/v1/deals/:id           # Delete deal
GET    /api/v1/deals               # List deals with pagination
```

### Pipeline & Analytics
```
GET    /api/v1/deals/pipeline      # Get pipeline overview with stages
GET    /api/v1/deals/owner/:id     # Get deals by owner ID
PUT    /api/v1/deals/:id/close     # Close a deal (won/lost)
```

### System
```
GET    /health                     # Health check endpoint
```

## Implemented Features

### Deal Lifecycle Management
- ✅ Complete deal CRUD operations
- ✅ Deal status tracking (open, won, lost)
- ✅ Deal closing workflows
- ✅ Owner assignment

### Pipeline Configuration
- Customizable pipeline stages
- Stage-specific probability settings
- Pipeline stage ordering
- Stage-based automation rules
- Pipeline performance metrics

### Revenue Forecasting
- Probability-weighted forecasting
- Time-based forecast projections
- Revenue recognition tracking
- Pipeline velocity analysis
- Win rate calculations

### Contact Integration
- Multiple contacts per deal
- Contact role assignments
- Primary contact designation
- Contact influence tracking
- Decision maker identification

### Sales Analytics
- Pipeline conversion rates
- Average deal size analysis
- Sales cycle duration
- Win/loss analysis
- Revenue trending
- Sales performance metrics

## Service Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://admin:admin@localhost:5433/crm-platform

# Currency Configuration
DEFAULT_CURRENCY=USD
SUPPORTED_CURRENCIES=USD,EUR,GBP,CAD

# Pipeline Configuration
DEFAULT_PIPELINE_STAGES=Lead,Qualified,Proposal,Negotiation,Closed Won,Closed Lost

# Analytics
FORECAST_PERIOD_MONTHS=6
MIN_DEAL_VALUE=0.01

# Application
PORT=8083
LOG_LEVEL=info
```

## Deal Pipeline Stages

### Default Pipeline Configuration
1. **Lead** (10% probability)
   - Initial opportunity identification
   - Basic qualification pending

2. **Qualified** (25% probability)
   - Budget and authority confirmed
   - Timeline established

3. **Proposal** (50% probability)
   - Formal proposal submitted
   - Solution presented

4. **Negotiation** (75% probability)
   - Terms being negotiated
   - Final decision pending

5. **Closed Won** (100% probability)
   - Deal successfully closed
   - Revenue recognized

6. **Closed Lost** (0% probability)
   - Deal lost to competitor or no decision
   - Loss reason documented

### Stage Progression Rules
- Deals can move forward or backward in stages
- Stage changes trigger probability updates
- Closed stages require additional information
- Automated notifications on stage changes

## Revenue Calculations

### Weighted Pipeline Value
```
Weighted Value = Deal Value × Probability × Currency Exchange Rate
```

### Forecast Calculations
- Monthly recurring revenue projections
- Quarterly forecast summaries
- Annual revenue targets
- Pipeline velocity metrics

### Win Rate Analysis
- Overall win rate percentage
- Win rate by stage
- Win rate by deal source
- Win rate by deal size

## Inter-Service Communication

### Outbound Calls (Planned)
- **Contact Service**: Contact and company information
- **Auth Service**: User authentication and context
- **Communication Service**: Deal activity logging
- **Email Service**: Deal notifications and updates

### Inbound Calls (Planned)
- **Contact Service**: Deal associations
- **Communication Service**: Deal-related activities
- **Frontend**: Deal management and analytics

### Event Publishing (Planned)
- `deal.created` - New opportunity opened
- `deal.stage_changed` - Pipeline progression
- `deal.closed_won` - Successful deal closure
- `deal.closed_lost` - Lost opportunity
- `deal.contact_added` - Contact association

## Performance Considerations

### Database Optimization
- Indexes on stage, status, created_at, expected_close_date
- Decimal type for precise financial calculations
- JSONB indexes for custom field searches
- Partitioning for large deal volumes

### Analytics Performance
- Materialized views for common analytics queries
- Caching for dashboard metrics
- Efficient aggregation queries
- Background calculation jobs

### Currency Handling
- Consistent decimal precision for money values
- Currency conversion rate caching
- Multi-currency reporting capabilities

## Security Considerations

### Data Protection
- Tenant isolation for all deal data
- Financial data encryption at rest
- Audit logging for deal modifications
- Revenue data access controls

### Access Control
- Role-based deal access (owner, team, read-only)
- Deal visibility rules
- Stage progression permissions
- Analytics access controls

## Testing Strategy

### Test Coverage
- ✅ Unit tests for handlers and business logic
- ✅ Integration tests for database operations
- ✅ End-to-end API tests
- ✅ Tenant isolation verification tests
- ✅ Performance benchmark tests

## Directory Structure

```
services/deal-service/
├── cmd/server/
│   ├── main.go                 # ✅ Fully implemented server
│   └── main_test.go           # Basic tests
├── internal/
│   ├── business/              # ✅ Business logic layer
│   ├── config/                # ✅ Configuration management
│   ├── db/                    # ✅ Generated SQLC code
│   ├── errors/                # ✅ Error definitions
│   ├── handlers/              # ✅ HTTP handlers
│   ├── middleware/            # ✅ Auth and tenant middleware
│   └── models/                # ✅ Request/response models
├── tests/
│   ├── e2e/                   # ✅ End-to-end tests
│   ├── integration/           # ✅ Integration tests
│   ├── unit/                  # ✅ Unit tests
│   ├── fixtures/              # Test data
│   └── helpers/               # Test utilities
├── db/
│   ├── queries/               # ✅ SQL query files
│   └── schema/                # ✅ Database schema
├── Dockerfile                 # ✅ Container definition
├── go.mod                     # ✅ Go dependencies
└── sqlc.yaml                  # ✅ SQLC configuration
```

## Related Documentation

- [Service Architecture](../architecture/services.md) - Overall microservices design
- [Database Design](../architecture/database.md) - Multi-tenant data architecture
- [SQLC Implementation](../architecture/sqlc.md) - Database query patterns
- [Contact Service](./contact-service.md) - Integration with contact management