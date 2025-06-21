# Deal Service

Sales pipeline and opportunity management service.

## Overview

The Deal Service manages sales opportunities through configurable pipeline stages, revenue forecasting, deal-contact associations, and comprehensive sales analytics for the MTenant CRM platform.

## Current Implementation Status

**Status**: SQLC configuration exists but generated code missing
- ✅ SQLC configuration (`sqlc.yaml`) 
- ✅ Database schema (`db/schema/`)
- ✅ SQL queries (`db/queries/`)
- ❌ Generated code (`internal/db/`) - **Needs `sqlc generate`**
- ❌ HTTP handlers and business logic (planned)
- ❌ Pipeline management logic (planned)
- ❌ Analytics and reporting (planned)

**Next Step**: Run `sqlc generate` in the deal service directory to create the generated database code.

## Database Schema

### Tenant-Specific Tables

**`deals`** - Sales opportunities with stage tracking
```sql
CREATE TABLE deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    value DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    stage VARCHAR(100) NOT NULL,
    probability INTEGER DEFAULT 0, -- 0-100 percentage
    expected_close_date DATE,
    actual_close_date DATE,
    status VARCHAR(50) DEFAULT 'open', -- open, won, lost
    company_id UUID, -- Reference to companies table in contact service
    contact_id UUID, -- Primary contact for the deal
    source VARCHAR(100), -- lead source
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    closed_by UUID,
    lost_reason VARCHAR(255)
);
```

**`deal_contacts`** - Many-to-many contact associations
```sql
CREATE TABLE deal_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID REFERENCES deals(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL, -- Reference to contacts in contact service
    role VARCHAR(100), -- decision_maker, influencer, user, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(deal_id, contact_id)
);
```

**`deal_stages`** - Configurable pipeline stages (planned)
```sql
CREATE TABLE deal_stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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

## Planned SQLC Queries

Once `sqlc generate` is run, the service will include:

- **Deal Management**: `CreateDeal`, `GetDealByID`, `UpdateDeal`, `DeleteDeal`, `ListDeals`
- **Pipeline Operations**: `GetDealsByStage`, `MoveDealToStage`, `GetPipelineOverview`
- **Deal Analytics**: `GetDealsByDateRange`, `GetRevenueByStage`, `GetWonDealsTotal`
- **Contact Associations**: `AddContactToDeal`, `RemoveContactFromDeal`, `GetDealContacts`
- **Forecasting**: `GetForecastData`, `GetDealsByExpectedCloseDate`

## Planned API Endpoints

**Current Status**: Endpoints not implemented - service has placeholder main.go

### Deal Management
```
POST   /api/deals                  # Create new deal
GET    /api/deals/:id              # Get deal details
PUT    /api/deals/:id              # Update deal
DELETE /api/deals/:id              # Soft delete deal
GET    /api/deals                  # List deals with filtering
```

### Pipeline Management
```
GET    /api/deals/pipeline         # Get pipeline overview
PUT    /api/deals/:id/stage        # Move deal to different stage
GET    /api/deals/stage/:stage     # Get deals by stage
POST   /api/deals/:id/notes        # Add deal notes
```

### Contact Associations
```
POST   /api/deals/:id/contacts     # Associate contact with deal
DELETE /api/deals/:id/contacts/:contact_id  # Remove contact association
GET    /api/deals/:id/contacts     # Get deal contacts
PUT    /api/deals/:id/contacts/:contact_id  # Update contact role
```

### Analytics & Reporting
```
GET    /api/deals/analytics        # Revenue analytics
GET    /api/deals/forecast         # Sales forecast data
GET    /api/deals/metrics          # Pipeline metrics
GET    /api/deals/reports          # Custom reports
```

## Planned Features

### Deal Lifecycle Management
- Complete deal CRUD operations
- Stage-based pipeline progression
- Automated probability calculations
- Deal status tracking (open, won, lost)
- Deal closing workflows

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

### Current Tests
- Basic Go module tests
- Utility function tests

### Planned Tests (After SQLC Generation)
- Deal CRUD operation tests
- Pipeline stage progression tests
- Revenue calculation validation tests
- Contact association tests
- Analytics and forecasting tests
- Currency conversion tests

## Development Status

**Current Directory Structure:**
```
services/deal-service/
├── cmd/server/
│   ├── main.go                 # Placeholder implementation
│   └── main_test.go           # Basic tests
├── internal/
│   ├── db/                    # MISSING - Needs sqlc generate
│   ├── benchmark_test.go      # Performance tests
│   └── utils_test.go          # Utility tests
├── db/
│   ├── queries/               # SQL query files ✅
│   └── schema/               # Database schema ✅
├── Dockerfile                 # Container definition
├── go.mod                    # Go dependencies ✅
└── sqlc.yaml                 # SQLC configuration ✅
```

## Immediate Next Steps

1. **Generate SQLC Code**: Run `sqlc generate` to create database access layer
2. **Verify Generated Code**: Test SQLC queries and database integration
3. **Implement Core Logic**: Build deal management business logic
4. **Create HTTP Handlers**: Develop REST API endpoints
5. **Add Pipeline Logic**: Implement stage management
6. **Build Analytics**: Create forecasting and reporting features

## Related Documentation

- [Service Architecture](../architecture/services.md) - Overall microservices design
- [Database Design](../architecture/database.md) - Multi-tenant data architecture
- [SQLC Implementation](../architecture/sqlc.md) - Database query patterns
- [Contact Service](./contact-service.md) - Integration with contact management