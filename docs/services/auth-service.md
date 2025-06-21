# Auth Service

User authentication, authorization, and token management service.

## Overview

The Auth Service handles user authentication, JWT token management, and role-based access control for the MTenant CRM platform. It serves as the central authentication authority for all other services.

## Current Implementation Status

**Status**: Basic SQLC setup completed, placeholder main.go implementation
- ✅ SQLC configuration (`sqlc.yaml`) 
- ✅ Database schema (`db/schema/`)
- ✅ SQL queries (`db/queries/`)
- ✅ Generated code (`internal/db/`)
- ❌ HTTP handlers and business logic (planned)
- ❌ JWT token implementation (planned)
- ❌ Password hashing and validation (planned)

## Database Schema

### Tables

**`users`** - User accounts with authentication data
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) DEFAULT 'user',
    tenant_id UUID REFERENCES tenants(id),
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**`password_reset_tokens`** - Temporary tokens for password recovery
```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## SQLC Queries

The service includes comprehensive SQLC queries for user management:

- **User Management**: `CreateUser`, `GetUserByEmail`, `GetUserByID`, `UpdateUser`, `DeleteUser`
- **Authentication**: `GetUserByEmailAndPassword`, `UpdatePassword`, `VerifyEmail`
- **Password Reset**: `CreatePasswordResetToken`, `GetPasswordResetToken`, `UsePasswordResetToken`

## Planned API Endpoints

**Current Status**: Endpoints not implemented - service has placeholder main.go

### Authentication Endpoints
```
POST   /api/auth/register          # User registration
POST   /api/auth/login             # User authentication
POST   /api/auth/refresh           # JWT token refresh
POST   /api/auth/logout            # Session termination
GET    /api/auth/profile           # Get user profile
PUT    /api/auth/profile           # Update user profile
```

### Password Management
```
POST   /api/auth/forgot-password   # Request password reset
POST   /api/auth/reset-password    # Confirm password reset
PUT    /api/auth/change-password   # Change password (authenticated)
```

### Token Validation
```
GET    /api/auth/validate          # Validate JWT token (for other services)
POST   /api/auth/verify-email      # Email verification
```

## Planned Features

### JWT Token Management
- Token generation with tenant context
- Refresh token rotation
- Token revocation support
- Secure cookie handling

### Password Security
- bcrypt password hashing
- Password strength validation
- Account lockout protection
- Rate limiting on login attempts

### Role-Based Access Control
- User roles: admin, user, read-only
- Tenant-scoped permissions
- Permission-based route protection
- Admin user management

## Service Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgresql://admin:admin@localhost:5433/crm-platform

# JWT Configuration
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY=1h
REFRESH_TOKEN_EXPIRY=7d

# Email Service (for verification/reset)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=noreply@example.com
SMTP_PASSWORD=smtp-password

# Application
PORT=8080
LOG_LEVEL=info
```

## Inter-Service Communication

### Outbound Calls (Planned)
- **Email Service**: Send verification and password reset emails
- **Tenant Service**: Validate tenant context during user registration

### Inbound Calls (Planned)
- **All Services**: Token validation requests
- **Frontend**: User authentication and profile management

## Security Considerations

### Authentication Security
- Password hashing with bcrypt (cost factor 12)
- JWT tokens with short expiry (1 hour)
- Secure refresh token rotation
- Rate limiting on authentication endpoints

### Data Protection
- No plaintext password storage
- Secure token generation
- Protected password reset flows
- Email verification requirements

## Testing Strategy

### Current Tests
- Basic SQLC generated code tests
- Database connection tests
- Utility function tests

### Planned Tests
- Authentication flow integration tests
- Password security validation tests
- JWT token generation and validation tests
- Rate limiting and security tests

## Development Status

**Current Directory Structure:**
```
services/auth-service/
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

1. **JWT Implementation**: Add JWT token generation and validation
2. **Password Security**: Implement bcrypt hashing and validation
3. **HTTP Handlers**: Create REST API endpoints
4. **Authentication Middleware**: Develop middleware for other services
5. **Email Integration**: Add email verification and password reset
6. **Rate Limiting**: Implement security controls
7. **Integration Testing**: Test with other services

## Related Documentation

- [Service Architecture](../architecture/services.md) - Overall microservices design
- [Database Design](../architecture/database.md) - Multi-tenant data architecture
- [SQLC Implementation](../architecture/sqlc.md) - Database query patterns
- [Auth Service Queries](../database/queries/auth-service.md) - SQL query documentation