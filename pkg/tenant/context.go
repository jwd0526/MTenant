package tenant

import (
    "context"
    "fmt"
    "regexp"
)

type contextKey string

var tenantKey contextKey = "tenantIDKey"

// UUID regex pattern for validation
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Context error definitions
var (
    ErrTenantNotFound    = fmt.Errorf("context does not contain tenant key")
    ErrInvalidTenantType = fmt.Errorf("tenant key must be string")
    ErrNilContext        = fmt.Errorf("cannot extract from nil context")
    ErrBlankTenantID     = fmt.Errorf("tenant ID cannot be blank")
    ErrInvalidTenantID   = fmt.Errorf("invalid tenant ID format")
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

// validateID validates tenant ID format (UUID)
func validateID(id string) error {
    if id == "" {
        return ErrBlankTenantID
    }

    if len(id) != 36 {
        return fmt.Errorf("%w: expected 36 characters, got %d", ErrInvalidTenantID, len(id))
    }

    if !uuidPattern.MatchString(id) {
        return fmt.Errorf("%w: must be valid UUID", ErrInvalidTenantID)
    }

    return nil
}