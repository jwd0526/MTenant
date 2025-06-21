# Communication Service

Customer interaction tracking and communication workflows service.

## Overview

The Communication Service manages all customer interactions including activity logging (emails, calls, meetings, notes), email sending with template support, interaction timelines, and task management for the MTenant CRM platform.

## Current Implementation Status

**Status**: No SQLC implementation - requires complete setup
- ❌ SQLC configuration (`sqlc.yaml`) - **Missing**
- ❌ Database schema (`db/schema/`) - **Missing**
- ❌ SQL queries (`db/queries/`) - **Missing**
- ❌ Generated code (`internal/db/`) - **Missing**
- ❌ HTTP handlers and business logic (planned)
- ❌ Email sending capabilities (planned)
- ❌ Activity tracking logic (planned)

**Next Steps**: Complete SQLC setup including schema, queries, and configuration before implementation can begin.

## Planned Database Schema

### Tenant-Specific Tables

**`activities`** - All customer interactions
```sql
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL, -- email, call, meeting, note, task
    subject VARCHAR(255),
    description TEXT,
    contact_id UUID, -- Reference to contacts in contact service
    deal_id UUID, -- Reference to deals in deal service
    company_id UUID, -- Reference to companies in contact service
    status VARCHAR(50) DEFAULT 'completed', -- completed, scheduled, cancelled
    direction VARCHAR(20), -- inbound, outbound
    duration_minutes INTEGER, -- for calls and meetings
    scheduled_at TIMESTAMP,
    completed_at TIMESTAMP,
    location VARCHAR(255), -- for meetings
    outcome VARCHAR(100), -- positive, negative, neutral, no_answer
    priority VARCHAR(20) DEFAULT 'normal', -- low, normal, high, urgent
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID
);
```

**`email_templates`** - Communication templates
```sql
CREATE TABLE email_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL,
    body_text TEXT,
    category VARCHAR(100), -- welcome, follow_up, proposal, etc.
    variables JSONB DEFAULT '{}', -- template variable definitions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);
```

**`email_tracking`** - Email delivery and engagement metrics
```sql
CREATE TABLE email_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID REFERENCES activities(id) ON DELETE CASCADE,
    email_address VARCHAR(255) NOT NULL,
    template_id UUID REFERENCES email_templates(id),
    message_id VARCHAR(255) UNIQUE, -- External email service message ID
    status VARCHAR(50) DEFAULT 'sent', -- sent, delivered, bounced, failed
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    bounced_at TIMESTAMP,
    bounce_reason VARCHAR(255),
    tracking_pixel_url VARCHAR(500),
    click_tracking JSONB DEFAULT '{}', -- clicked links and timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**`tasks`** - Task and reminder management
```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'task', -- task, reminder, follow_up
    status VARCHAR(50) DEFAULT 'pending', -- pending, in_progress, completed, cancelled
    priority VARCHAR(20) DEFAULT 'normal', -- low, normal, high, urgent
    due_date TIMESTAMP,
    reminder_at TIMESTAMP,
    contact_id UUID, -- Reference to contacts in contact service
    deal_id UUID, -- Reference to deals in deal service
    assigned_to UUID, -- User ID
    completed_at TIMESTAMP,
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    completed_by UUID
);
```

## Planned SQLC Queries

Once SQLC setup is complete, the service will include:

- **Activity Management**: `CreateActivity`, `GetActivityByID`, `UpdateActivity`, `DeleteActivity`, `ListActivities`
- **Activity Search**: `GetActivitiesByContact`, `GetActivitiesByDeal`, `GetActivitiesByDateRange`
- **Email Templates**: `CreateEmailTemplate`, `GetEmailTemplate`, `ListEmailTemplates`, `UpdateEmailTemplate`
- **Email Tracking**: `CreateEmailTracking`, `UpdateEmailDeliveryStatus`, `TrackEmailOpen`, `TrackEmailClick`
- **Task Management**: `CreateTask`, `GetTaskByID`, `UpdateTask`, `CompleteTask`, `GetTasksByUser`
- **Timeline**: `GetContactTimeline`, `GetDealTimeline`, `GetActivityTimeline`

## Planned API Endpoints

**Current Status**: Endpoints not implemented - service has placeholder main.go

### Activity Management
```
POST   /api/activities             # Log new activity
GET    /api/activities/:id         # Get activity details
PUT    /api/activities/:id         # Update activity
DELETE /api/activities/:id         # Delete activity
GET    /api/activities             # List activities with filtering
```

### Activity Timeline
```
GET    /api/activities/contact/:id     # Get contact activity timeline
GET    /api/activities/deal/:id        # Get deal activity timeline
GET    /api/activities/company/:id     # Get company activity timeline
GET    /api/activities/timeline        # Get combined timeline
```

### Email Management
```
POST   /api/emails/send            # Send email (with optional template)
GET    /api/emails/templates       # List email templates
POST   /api/emails/templates       # Create email template
PUT    /api/emails/templates/:id   # Update email template
GET    /api/emails/tracking/:id    # Get email tracking data
```

### Task Management
```
POST   /api/tasks                  # Create task/reminder
GET    /api/tasks/:id              # Get task details
PUT    /api/tasks/:id              # Update task
DELETE /api/tasks/:id              # Delete task
GET    /api/tasks                  # List tasks (mine/assigned)
POST   /api/tasks/:id/complete     # Mark task as completed
```

### Analytics & Reporting
```
GET    /api/analytics/activities   # Activity analytics
GET    /api/analytics/emails       # Email performance metrics
GET    /api/analytics/engagement   # Contact engagement scores
```

## Planned Features

### Activity Tracking
- Comprehensive activity logging (emails, calls, meetings, notes)
- Activity categorization and tagging
- Time tracking for calls and meetings
- Activity outcomes and follow-up actions
- Custom field support for activity details

### Email Management
- Template-based email sending
- Variable substitution in templates
- HTML and plain text email support
- Email delivery tracking
- Open and click tracking
- Bounce and failure handling

### Communication Analytics
- Email engagement metrics
- Response rate tracking
- Communication frequency analysis
- Contact engagement scoring
- Activity timeline visualization

### Task & Reminder System
- Task creation and assignment
- Due date and reminder management
- Task prioritization
- Task completion tracking
- Follow-up task automation

### Timeline & History
- Unified activity timeline
- Contact interaction history
- Deal communication tracking
- Cross-service activity aggregation

## Service Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://admin:admin@localhost:5433/crm-platform

# Email Service Provider
EMAIL_PROVIDER=sendgrid # or ses, mailgun, smtp
SENDGRID_API_KEY=your-sendgrid-api-key
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=noreply@example.com
SMTP_PASSWORD=smtp-password

# Email Tracking
TRACKING_DOMAIN=track.example.com
TRACKING_PIXEL_ENABLED=true
CLICK_TRACKING_ENABLED=true

# Task Management
REMINDER_CHECK_INTERVAL=5m
TASK_AUTO_ASSIGNMENT=true

# Application
PORT=8084
LOG_LEVEL=info
```

## Email Template System

### Template Variables
Templates support dynamic variable substitution:

```html
<!-- Email template example -->
<h1>Hello {{.contact.first_name}}!</h1>
<p>Thank you for your interest in {{.deal.title}}.</p>
<p>Your {{.company.name}} account is now active.</p>
<p>Next steps: {{.custom.next_steps}}</p>
```

### Variable Categories
- **Contact Variables**: `{{.contact.first_name}}`, `{{.contact.email}}`, etc.
- **Company Variables**: `{{.company.name}}`, `{{.company.industry}}`, etc.
- **Deal Variables**: `{{.deal.title}}`, `{{.deal.value}}`, etc.
- **Custom Variables**: User-defined template variables
- **System Variables**: `{{.current_date}}`, `{{.user.name}}`, etc.

## Email Tracking Features

### Delivery Tracking
- Sent confirmation
- Delivery confirmation
- Bounce detection and categorization
- Spam folder detection

### Engagement Tracking
- Email open tracking (pixel-based)
- Link click tracking
- Time-based engagement metrics
- Device and client detection

### Analytics Data
- Open rates by template
- Click-through rates
- Best performing email times
- Engagement scoring

## Task Management Features

### Task Types
- **Tasks**: General work items
- **Reminders**: Time-based notifications
- **Follow-ups**: Contact or deal-based actions
- **Appointments**: Scheduled meetings or calls

### Task Assignment
- Self-assigned tasks
- Team task assignment
- Automatic task creation from activities
- Task delegation and ownership transfer

### Reminder System
- Email notifications
- In-app notifications
- SMS reminders (planned)
- Escalation workflows

## Inter-Service Communication

### Outbound Calls (Planned)
- **Contact Service**: Contact and company information
- **Deal Service**: Deal details and associations
- **Auth Service**: User authentication and context
- **Email Provider APIs**: Email sending and tracking

### Inbound Calls (Planned)
- **All Services**: Activity logging requests
- **Frontend**: Communication management interface
- **Webhook Endpoints**: Email provider status updates

### Event Consumption (Planned)
- `contact.created` - Send welcome email
- `deal.closed_won` - Trigger celebration email
- `deal.stage_changed` - Log automatic activity
- `user.created` - Send onboarding email sequence

## Performance Considerations

### Activity Storage
- Efficient indexing for timeline queries
- Pagination for large activity volumes
- Archive strategy for old activities
- Bulk activity import capabilities

### Email Processing
- Asynchronous email sending
- Queue-based email delivery
- Rate limiting for email providers
- Retry logic for failed sends

### Analytics Performance
- Materialized views for common metrics
- Cached engagement scores
- Background metric calculations
- Efficient timeline aggregation

## Security Considerations

### Email Security
- SPF, DKIM, and DMARC configuration
- Secure email template rendering
- Link safety validation
- Tracking pixel privacy compliance

### Data Protection
- Tenant isolation for all communication data
- Activity data encryption
- Email content encryption at rest
- GDPR compliance for communication tracking

### Access Control
- Activity visibility permissions
- Email template access control
- Task assignment permissions
- Communication data export controls

## Development Status

**Current Directory Structure:**
```
services/communication-service/
├── cmd/server/
│   ├── main.go                 # Placeholder implementation
│   └── main_test.go           # Basic tests
├── internal/
│   ├── benchmark_test.go      # Performance tests
│   └── utils_test.go          # Utility tests
├── db/                        # MISSING - Needs complete setup
├── Dockerfile                 # Container definition
├── go.mod                    # Go dependencies
└── sqlc.yaml                 # MISSING - Needs creation
```

## Immediate Next Steps

1. **Create SQLC Configuration**: Set up `sqlc.yaml` with email and activity configurations
2. **Design Database Schema**: Create comprehensive schema for activities, emails, and tasks
3. **Write SQL Queries**: Develop SQLC queries for all operations
4. **Generate SQLC Code**: Run `sqlc generate` to create database access layer
5. **Implement Email Provider Integration**: Add email sending capabilities
6. **Build Activity Tracking**: Implement activity logging and timeline features
7. **Create Task Management**: Build task and reminder system

## Related Documentation

- [Service Architecture](../architecture/services.md) - Overall microservices design
- [Database Design](../architecture/database.md) - Multi-tenant data architecture
- [SQLC Implementation](../architecture/sqlc.md) - Database query patterns
- [Contact Service](./contact-service.md) - Integration with contact management
- [Deal Service](./deal-service.md) - Integration with deal management