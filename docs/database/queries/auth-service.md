# Auth Service SQL Queries

Complete documentation of SQL queries used by the Auth Service for user authentication and management.

## Overview

The Auth Service manages user authentication, password resets, and session management within tenant schemas. All queries operate within the tenant context set by middleware.

## User Management Queries

### User Creation

**Query: `CreateUser`**
```sql
-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, role, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, first_name, last_name, role, active, email_verified, created_at;
```

**Purpose:** Create a new user account within the tenant schema.

**Parameters:**
- `$1` - Email address (VARCHAR 254)
- `$2` - Hashed password (VARCHAR 60, bcrypt)
- `$3` - First name (VARCHAR 100)
- `$4` - Last name (VARCHAR 100)
- `$5` - Role ('admin', 'manager', 'sales_rep', 'viewer')
- `$6` - Created by user ID (nullable)

**Generated Go Method:**
```go
type CreateUserParams struct {
    Email        string `json:"email"`
    PasswordHash string `json:"password_hash"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Role         string `json:"role"`
    CreatedBy    *int32 `json:"created_by"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error)
```

**Usage Example:**
```go
user, err := queries.CreateUser(ctx, db.CreateUserParams{
    Email:        "john.doe@example.com",
    PasswordHash: "$2a$10$...",
    FirstName:    "John",
    LastName:     "Doe",
    Role:         "sales_rep",
    CreatedBy:    &adminUserID,
})
```

### User Authentication

**Query: `GetUserForAuth`**
```sql
-- name: GetUserForAuth :one
SELECT id, password_hash, role, active, email_verified
FROM users 
WHERE email = $1 AND active = true AND deleted_at IS NULL;
```

**Purpose:** Retrieve minimal user data required for authentication validation.

**Parameters:**
- `$1` - Email address

**Security Features:**
- Only returns active, non-deleted users
- Includes only necessary fields for auth
- Password hash for bcrypt comparison

**Usage Example:**
```go
authUser, err := queries.GetUserForAuth(ctx, "john.doe@example.com")
if err != nil {
    return errors.New("invalid credentials")
}

if !bcrypt.CompareHashAndPassword([]byte(authUser.PasswordHash), []byte(password)) {
    return errors.New("invalid credentials")
}
```

### User Profile Retrieval

**Query: `GetUserByEmail`**
```sql
-- name: GetUserByEmail :one
SELECT id, email, password_hash, first_name, last_name, role, permissions, 
       active, email_verified, last_login, created_at, updated_at
FROM users
WHERE email = $1 AND active = true AND deleted_at IS NULL;
```

**Query: `GetUserByID`**
```sql
-- name: GetUserByID :one
SELECT id, email, first_name, last_name, role, permissions, 
       active, email_verified, last_login, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;
```

**Purpose:** Get complete user profile data for API responses.

**Usage Example:**
```go
// Get user profile
user, err := queries.GetUserByID(ctx, userID)
if err != nil {
    return c.JSON(404, gin.H{"error": "User not found"})
}

// Remove sensitive data before response
response := UserProfileResponse{
    ID:            user.ID,
    Email:         user.Email,
    FirstName:     user.FirstName,
    LastName:      user.LastName,
    Role:          user.Role,
    EmailVerified: user.EmailVerified,
    LastLogin:     user.LastLogin,
}
```

### Session Management

**Query: `UpdateUserLastLogin`**
```sql
-- name: UpdateUserLastLogin :exec
UPDATE users 
SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
```

**Purpose:** Record user login timestamp for session tracking.

**Usage Example:**
```go
// Update login timestamp after successful authentication
err := queries.UpdateUserLastLogin(ctx, user.ID)
if err != nil {
    log.Printf("Failed to update last login for user %d: %v", user.ID, err)
    // Non-critical error, continue with login
}
```

### Email Verification

**Query: `VerifyUserEmail`**
```sql
-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = true, updated_at = CURRENT_TIMESTAMP, updated_by = $2
WHERE id = $1;
```

**Purpose:** Mark user email as verified after confirmation.

**Parameters:**
- `$1` - User ID
- `$2` - Admin/system user performing verification

**Usage Example:**
```go
// Verify email after token validation
err := queries.VerifyUserEmail(ctx, db.VerifyUserEmailParams{
    ID:        userID,
    UpdatedBy: &systemUserID,
})
```

## User Management Operations

### Role and Permission Updates

**Query: `UpdateUserRole`**
```sql
-- name: UpdateUserRole :exec
UPDATE users
SET role = $2, updated_at = CURRENT_TIMESTAMP, updated_by = $3
WHERE id = $1 AND deleted_at IS NULL;
```

**Query: `UpdateUserPermissions`**
```sql
-- name: UpdateUserPermissions :exec
UPDATE users
SET permissions = $2, updated_at = CURRENT_TIMESTAMP, updated_by = $3
WHERE id = $1 AND deleted_at IS NULL;
```

**Purpose:** Administrative functions for role-based access control.

**Usage Example:**
```go
// Promote user to manager role
err := queries.UpdateUserRole(ctx, db.UpdateUserRoleParams{
    ID:        userID,
    Role:      "manager",
    UpdatedBy: &adminUserID,
})

// Grant specific permissions
permissions := map[string]bool{
    "view_analytics": true,
    "export_data":   true,
    "manage_deals":  false,
}
permissionJSON, _ := json.Marshal(permissions)

err = queries.UpdateUserPermissions(ctx, db.UpdateUserPermissionsParams{
    ID:          userID,
    Permissions: permissionJSON,
    UpdatedBy:   &adminUserID,
})
```

### Password Management

**Query: `UpdateUserPassword`**
```sql
-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = CURRENT_TIMESTAMP, updated_by = $3
WHERE id = $1 AND deleted_at IS NULL;
```

**Purpose:** Update user password with new hash.

**Usage Example:**
```go
// Hash new password
newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
if err != nil {
    return err
}

// Update password in database
err = queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
    ID:           userID,
    PasswordHash: string(newHash),
    UpdatedBy:    &userID, // Self-update
})
```

### User Deactivation

**Query: `DeactivateUser`**
```sql
-- name: DeactivateUser :exec
UPDATE users
SET active = false, updated_at = CURRENT_TIMESTAMP, updated_by = $2
WHERE id = $1 AND deleted_at IS NULL;
```

**Query: `SoftDeleteUser`**
```sql
-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP, updated_by = $2
WHERE id = $1;
```

**Purpose:** Deactivate or soft-delete users while preserving data integrity.

**Usage Example:**
```go
// Temporarily deactivate user
err := queries.DeactivateUser(ctx, db.DeactivateUserParams{
    ID:        userID,
    UpdatedBy: &adminUserID,
})

// Permanently soft-delete user (preserves foreign keys)
err = queries.SoftDeleteUser(ctx, db.SoftDeleteUserParams{
    ID:        userID,
    UpdatedBy: &adminUserID,
})
```

## User Listing and Search

### Active User Lists

**Query: `ListActiveUsers`**
```sql
-- name: ListActiveUsers :many
SELECT id, email, first_name, last_name, role, active, email_verified, 
       last_login, created_at
FROM users
WHERE active = true AND deleted_at IS NULL
ORDER BY first_name, last_name;
```

**Query: `ListUsersByRole`**
```sql
-- name: ListUsersByRole :many
SELECT id, email, first_name, last_name, role, active, email_verified,
       last_login, created_at
FROM users
WHERE role = $1 AND active = true AND deleted_at IS NULL
ORDER BY first_name, last_name;
```

**Purpose:** Administrative user management and role filtering.

**Usage Example:**
```go
// Get all active users for admin panel
users, err := queries.ListActiveUsers(ctx)

// Get all sales reps for assignment
salesReps, err := queries.ListUsersByRole(ctx, "sales_rep")
```

### Email Validation

**Query: `CheckEmailExists`**
```sql
-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE email = $1 AND deleted_at IS NULL
);
```

**Purpose:** Validate email uniqueness before user creation.

**Usage Example:**
```go
exists, err := queries.CheckEmailExists(ctx, "new.user@example.com")
if err != nil {
    return err
}
if exists {
    return errors.New("email already in use")
}
```

## Password Reset Tokens

### Token Creation

**Query: `CreatePasswordResetToken`**
```sql
-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token, expires_at, created_at;
```

**Purpose:** Generate secure token for password reset workflow.

**Usage Example:**
```go
// Generate secure random token
token := generateSecureToken(32)
expiry := time.Now().Add(time.Hour * 24) // 24-hour expiry

resetToken, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
    UserID:    &userID,
    Token:     token,
    ExpiresAt: expiry,
})
```

### Token Validation

**Query: `GetPasswordResetToken`**
```sql
-- name: GetPasswordResetToken :one
SELECT prt.id, prt.user_id, prt.token, prt.expires_at, prt.used_at,
       u.email, u.first_name, u.last_name
FROM password_reset_tokens prt
JOIN users u ON prt.user_id = u.id
WHERE prt.token = $1 AND prt.expires_at > CURRENT_TIMESTAMP;
```

**Purpose:** Validate reset token and get associated user data.

**Usage Example:**
```go
resetData, err := queries.GetPasswordResetToken(ctx, token)
if err != nil {
    return errors.New("invalid or expired token")
}

if resetData.UsedAt.Valid {
    return errors.New("token already used")
}
```

### Token Usage

**Query: `MarkPasswordResetTokenUsed`**
```sql
-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens 
SET used_at = CURRENT_TIMESTAMP
WHERE token = $1;
```

**Purpose:** Mark token as used to prevent reuse.

**Usage Example:**
```go
// After successful password reset
err := queries.MarkPasswordResetTokenUsed(ctx, token)
if err != nil {
    log.Printf("Failed to mark token as used: %v", err)
    // Continue - password was already reset
}
```

## Query Performance Notes

### Index Usage

**Optimized Queries:**
- Email lookups use `idx_users_email` index
- Active user filters use `idx_users_active` index
- Role filtering uses `idx_users_role` index
- Token lookups use `idx_password_reset_tokens_token` index

**Query Plans:**
```sql
-- Verify index usage for critical auth query
EXPLAIN (ANALYZE, BUFFERS) 
SELECT id, password_hash, role, active, email_verified
FROM users 
WHERE email = 'john.doe@example.com' AND active = true AND deleted_at IS NULL;

-- Should show: Index Scan using idx_users_email
```

### Performance Considerations

**Fast Authentication:**
- `GetUserForAuth` returns minimal fields
- Email index ensures O(log n) lookup
- Active/deleted filters use compound conditions

**Session Management:**
- `UpdateUserLastLogin` is async-safe
- Minimal lock time on user row
- Non-blocking for concurrent requests

**Security Features:**
- All queries respect soft-delete patterns
- Password hashes never returned in profile queries
- Token expiry checked at database level

## Error Handling Patterns

### Common Errors

**User Not Found:**
```go
user, err := queries.GetUserByEmail(ctx, email)
if errors.Is(err, pgx.ErrNoRows) {
    return c.JSON(404, gin.H{"error": "User not found"})
}
```

**Constraint Violations:**
```go
_, err := queries.CreateUser(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        return c.JSON(409, gin.H{"error": "Email already in use"})
    }
    return c.JSON(500, gin.H{"error": "Database error"})
}
```

**Token Validation:**
```go
resetToken, err := queries.GetPasswordResetToken(ctx, token)
if errors.Is(err, pgx.ErrNoRows) {
    return c.JSON(400, gin.H{"error": "Invalid or expired token"})
}
```

## Related Documentation

- [User Schema](../tenant-template/UserSchema.md) - Complete table definitions
- [SQLC Configuration](../../architecture/sqlc.md) - Code generation details
- [Auth Service Architecture](../../services/auth-service.md) - Service implementation