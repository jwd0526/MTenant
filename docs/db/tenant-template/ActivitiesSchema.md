# Activities Table Documentation

## Purpose
Customer interaction tracking within tenant schemas. Stores all communications, tasks, and activities related to contacts, companies, and deals for complete relationship history and follow-up management.

## Table Structure
```sql
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('email', 'call', 'meeting', 'note', 'task')),
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Polymorphic relationships (related_type + related_id)
    related_type VARCHAR(20) CHECK (related_type IN ('contact', 'company', 'deal')),
    related_id INTEGER,
    
    -- Task-specific fields
    due_date TIMESTAMP,
    completed_at TIMESTAMP,
    is_completed BOOLEAN DEFAULT FALSE,
    priority VARCHAR(10) CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    
    -- Call/Meeting specific fields
    duration_minutes INTEGER, -- Duration in minutes
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    location VARCHAR(255), -- Meeting location or call number
    
    -- Email specific fields
    email_direction VARCHAR(10) CHECK (email_direction IN ('inbound', 'outbound')),
    email_status VARCHAR(20) CHECK (email_status IN ('sent', 'delivered', 'opened', 'clicked', 'bounced')),
    
    -- Participants (for meetings/calls with multiple people)
    participants JSONB, -- Array of contact IDs or external participants
    
    -- File attachments
    attachments JSONB, -- Array of file references
    
    -- Ownership and audit
    owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id)
);

-- Indexes for performance
CREATE INDEX idx_activities_type ON activities(type);
CREATE INDEX idx_activities_related ON activities(related_type, related_id);
CREATE INDEX idx_activities_owner ON activities(owner_id);
CREATE INDEX idx_activities_due_date ON activities(due_date);
CREATE INDEX idx_activities_created_at ON activities(created_at);
CREATE INDEX idx_activities_completed ON activities(is_completed, due_date);

-- Composite index for timeline queries
CREATE INDEX idx_activities_timeline ON activities(related_type, related_id, created_at DESC);

-- Partial index for incomplete tasks
CREATE INDEX idx_activities_open_tasks ON activities(owner_id, due_date) 
WHERE type = 'task' AND is_completed = FALSE;
```

## Column Design Decisions
- **type**: Enumerated activity types for consistent categorization and filtering
- **related_type/related_id**: Polymorphic relationship allows linking to contacts, companies, or deals
- **description**: TEXT field supports rich content and detailed notes
- **duration_minutes**: Simple integer for call/meeting duration tracking
- **participants**: JSONB array for flexible participant storage (internal users + external contacts)
- **attachments**: JSONB array for file references and metadata
- **priority**: Task prioritization for workload management
- **email_status**: Tracks email engagement for sales intelligence

## Schema Context
- Lives within tenant schemas for complete data isolation
- Template copied from `tenant_template` to each new tenant schema
- Polymorphic design allows single table for all activity types
- Referenced by timeline views across contacts, companies, and deals

## Key Operations
```sql
-- Log a phone call
INSERT INTO activities (type, subject, description, related_type, related_id, duration_minutes, start_time, owner_id, created_by)
VALUES ('call', 'Discovery call with John Smith', 'Discussed requirements for new CRM system. Interested in enterprise features.', 'contact', 1, 45, '2024-01-15 10:00:00', 1, 1);

-- Create a follow-up task
INSERT INTO activities (type, subject, description, related_type, related_id, due_date, priority, owner_id, created_by)
VALUES ('task', 'Send proposal to TechFlow', 'Prepare and send detailed proposal based on discovery call requirements', 'deal', 1, '2024-01-20 17:00:00', 'high', 1, 1);

-- Log an email with tracking
INSERT INTO activities (type, subject, description, related_type, related_id, email_direction, email_status, owner_id, created_by)
VALUES ('email', 'Re: CRM Implementation Timeline', 'Sent updated project timeline and next steps', 'deal', 1, 'outbound', 'sent', 1, 1);

-- Schedule a meeting with multiple participants
INSERT INTO activities (type, subject, description, related_type, related_id, start_time, end_time, location, participants, owner_id, created_by)
VALUES ('meeting', 'Technical Requirements Review', 'Review technical requirements with IT team', 'deal', 1, '2024-01-18 14:00:00', '2024-01-18 15:30:00', 'Conference Room A', '[{"type": "contact", "id": 1}, {"type": "external", "name": "Jane Doe", "email": "jane@techflow.com"}]', 1, 1);
```

## Polymorphic Relationship Queries
```sql
-- Get all activities for a specific contact
SELECT * FROM activities 
WHERE related_type = 'contact' AND related_id = 1 
ORDER BY created_at DESC;

-- Get timeline for a deal (all related activities)
SELECT a.*, c.first_name, c.last_name
FROM activities a
LEFT JOIN contacts c ON a.related_type = 'contact' AND a.related_id = c.id
WHERE (a.related_type = 'deal' AND a.related_id = 1)
   OR (a.related_type = 'contact' AND a.related_id IN (
       SELECT primary_contact_id FROM deals WHERE id = 1
   ))
ORDER BY a.created_at DESC;

-- Get all activities across multiple entities
SELECT 
    a.*,
    CASE 
        WHEN a.related_type = 'contact' THEN c.first_name || ' ' || c.last_name
        WHEN a.related_type = 'company' THEN comp.name
        WHEN a.related_type = 'deal' THEN d.title
    END as related_name
FROM activities a
LEFT JOIN contacts c ON a.related_type = 'contact' AND a.related_id = c.id
LEFT JOIN companies comp ON a.related_type = 'company' AND a.related_id = comp.id
LEFT JOIN deals d ON a.related_type = 'deal' AND a.related_id = d.id
WHERE a.owner_id = 1
ORDER BY a.created_at DESC;
```

## Task Management Queries
```sql
-- Get overdue tasks
SELECT * FROM activities 
WHERE type = 'task' 
  AND is_completed = FALSE 
  AND due_date < NOW()
ORDER BY due_date ASC;

-- Complete a task
UPDATE activities 
SET is_completed = TRUE, completed_at = NOW(), updated_at = NOW()
WHERE id = 1 AND type = 'task';

-- Daily task summary
SELECT 
    priority,
    COUNT(*) as task_count,
    COUNT(CASE WHEN is_completed THEN 1 END) as completed_count
FROM activities 
WHERE type = 'task' 
  AND DATE(due_date) = CURRENT_DATE
  AND owner_id = 1
GROUP BY priority;
```

## Activity Analytics
```sql
-- Communication frequency by type
SELECT type, COUNT(*) as activity_count, AVG(duration_minutes) as avg_duration
FROM activities 
WHERE created_at >= NOW() - INTERVAL '30 days'
  AND type IN ('call', 'meeting', 'email')
GROUP BY type;

-- Sales rep activity summary
SELECT 
    u.first_name, 
    u.last_name,
    COUNT(*) as total_activities,
    COUNT(CASE WHEN a.type = 'call' THEN 1 END) as calls,
    COUNT(CASE WHEN a.type = 'email' THEN 1 END) as emails,
    AVG(CASE WHEN a.type = 'call' THEN a.duration_minutes END) as avg_call_duration
FROM activities a
JOIN users u ON a.owner_id = u.id
WHERE a.created_at >= NOW() - INTERVAL '7 days'
GROUP BY u.id, u.first_name, u.last_name;
```

## JSONB Usage Examples
```sql
-- Store meeting participants
UPDATE activities SET participants = '[
    {"type": "user", "id": 1, "name": "Sales Rep"},
    {"type": "contact", "id": 2, "name": "John Smith"},
    {"type": "external", "name": "Jane Doe", "email": "jane@techflow.com", "title": "CTO"}
]' WHERE id = 1;

-- Store file attachments
UPDATE activities SET attachments = '[
    {"filename": "proposal.pdf", "size": 1024000, "url": "/files/proposal.pdf"},
    {"filename": "contract.docx", "size": 512000, "url": "/files/contract.docx"}
]' WHERE id = 1;

-- Query activities with external participants
SELECT * FROM activities 
WHERE participants @> '[{"type": "external"}]';
```

## Application Flow
1. Sales rep logs activity through CRM interface
2. System determines polymorphic relationship based on context
3. Activity appears in relevant timelines (contact, company, deal)
4. Tasks generate reminders and appear in to-do lists
5. Email activities update engagement tracking
6. Analytics aggregate activity data for performance reporting
7. All queries filtered by tenant schema for data isolation