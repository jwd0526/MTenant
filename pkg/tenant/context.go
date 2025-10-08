package tenant

import (
    "context"
    "fmt"
    
    "github.com/oklog/ulid/v2"
)

type contextKey string

var tenantKey contextKey = "tenantIDKey"

// ULID validation using oklog/ulid library

// Context error definitions
var (
    ErrTenantNotFound    = fmt.Errorf("context does not contain tenant key")
    ErrInvalidTenantType = fmt.Errorf("tenant key must be string")
    ErrNilContext        = fmt.Errorf("cannot extract from nil context")
    ErrBlankTenantID     = fmt.Errorf("tenant ID cannot be blank")
    ErrInvalidTenantID   = fmt.Errorf("invalid ULID format")
)

// NewContext creates a new context with tenant ID stored
func NewContext(ctx context.Context, tenantID string) (context.Context, error) {
    if err := validateID(tenantID); err != nil {
        return nil, fmt.Errorf("failed to create tenant context: %w", err)
    }
    return context.WithValue(ctx, tenantKey, tenantID), nil
}

// FromContext extracts tenant ID from context
func FromContext(ctx context.Context) (string, error) {
    if ctx == nil {
        return "", ErrNilContext
    }

    v := ctx.Value(tenantKey)
    if v == nil {
        return "", ErrTenantNotFound
    }

    value, ok := v.(string)
    if !ok {
        return "", fmt.Errorf("%w: incorrect type: %T", ErrInvalidTenantType, v)
    }

    return value, nil
}

// HasTenant checks if context contains a tenant ID
func HasTenant(ctx context.Context) bool {
    if ctx == nil {
        return false
    }
    _, ok := ctx.Value(tenantKey).(string)
    return ok
}

// MustFromContext extracts tenant ID or panics - use only when tenant is guaranteed
func MustFromContext(ctx context.Context) string {
    tenantID, err := FromContext(ctx)
    if err != nil {
        panic(fmt.Sprintf("tenant context required: %v", err))
    }
    return tenantID
}

// validateID validates tenant ID format (ULID)
func validateID(id string) error {
    if id == "" {
        return ErrBlankTenantID
    }

    if len(id) != 26 {
        return fmt.Errorf("%w: expected 26 characters, got %d", ErrInvalidTenantID, len(id))
    }

    _, err := ulid.Parse(id)
    if err != nil {
        return fmt.Errorf("%w: %v", ErrInvalidTenantID, err)
    }

    return nil
}