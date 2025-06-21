package tenant

import (
	"context"
	"testing"
)

func TestNewContext(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		wantError bool
	}{
		{
			name:      "valid UUID",
			tenantID:  "550e8400-e29b-41d4-a716-446655440000",
			wantError: false,
		},
		{
			name:      "empty tenant ID",
			tenantID:  "",
			wantError: true,
		},
		{
			name:      "invalid length",
			tenantID:  "short",
			wantError: true,
		},
		{
			name:      "invalid UUID format",
			tenantID:  "550e8400-e29b-41d4-a716-44665544000g",
			wantError: true,
		},
		{
			name:      "missing hyphens",
			tenantID:  "550e8400e29b41d4a716446655440000",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx, err := NewContext(ctx, tt.tenantID)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewContext() expected error but got none")
				}
				if newCtx != nil {
					t.Errorf("NewContext() expected nil context on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewContext() unexpected error: %v", err)
				}
				if newCtx == nil {
					t.Errorf("NewContext() returned nil context")
				}
			}
		})
	}
}

func TestFromContext(t *testing.T) {
	validTenantID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("nil context", func(t *testing.T) {
		_, err := FromContext(context.TODO())
		if err == nil {
			t.Errorf("FromContext() with nil context should return error")
		}
	})

	t.Run("empty context", func(t *testing.T) {
		_, err := FromContext(context.TODO())
		if err == nil {
			t.Errorf("FromContext() with empty context should return error")
		}
	})

	t.Run("valid tenant in context", func(t *testing.T) {
		ctx := context.Background()
		newCtx, err := NewContext(ctx, validTenantID)
		if err != nil {
			t.Fatalf("NewContext() failed: %v", err)
		}

		tenantID, err := FromContext(newCtx)
		if err != nil {
			t.Errorf("FromContext() unexpected error: %v", err)
		}
		if tenantID != validTenantID {
			t.Errorf("FromContext() = %v, want %v", tenantID, validTenantID)
		}
	})

	t.Run("wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), tenantKey, 123)
		_, err := FromContext(ctx)
		if err == nil {
			t.Errorf("FromContext() with wrong type should return error")
		}
	})
}

func TestHasTenant(t *testing.T) {
	validTenantID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("nil context", func(t *testing.T) {
		if HasTenant(context.TODO()) {
			t.Errorf("HasTenant() with nil context should return false")
		}
	})

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		if HasTenant(ctx) {
			t.Errorf("HasTenant() with empty context should return false")
		}
	})

	t.Run("context with tenant", func(t *testing.T) {
		ctx := context.Background()
		newCtx, err := NewContext(ctx, validTenantID)
		if err != nil {
			t.Fatalf("NewContext() failed: %v", err)
		}

		if !HasTenant(newCtx) {
			t.Errorf("HasTenant() should return true for context with tenant")
		}
	})

	t.Run("wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), tenantKey, 123)
		if HasTenant(ctx) {
			t.Errorf("HasTenant() with wrong type should return false")
		}
	})
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		wantError bool
	}{
		{
			name:      "valid UUID",
			tenantID:  "550e8400-e29b-41d4-a716-446655440000",
			wantError: false,
		},
		{
			name:      "valid UUID lowercase",
			tenantID:  "550e8400-e29b-41d4-a716-446655440000",
			wantError: false,
		},
		{
			name:      "valid UUID uppercase",
			tenantID:  "550E8400-E29B-41D4-A716-446655440000",
			wantError: false,
		},
		{
			name:      "empty string",
			tenantID:  "",
			wantError: true,
		},
		{
			name:      "too short",
			tenantID:  "550e8400-e29b",
			wantError: true,
		},
		{
			name:      "too long",
			tenantID:  "550e8400-e29b-41d4-a716-446655440000-extra",
			wantError: true,
		},
		{
			name:      "invalid characters",
			tenantID:  "550e8400-e29b-41d4-a716-44665544000g",
			wantError: true,
		},
		{
			name:      "missing hyphens",
			tenantID:  "550e8400e29b41d4a716446655440000",
			wantError: true,
		},
		{
			name:      "wrong hyphen positions",
			tenantID:  "550e840-0e29b-41d4-a716-446655440000",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateID(tt.tenantID)

			if tt.wantError {
				if err == nil {
					t.Errorf("validateID() expected error but got none for input: %s", tt.tenantID)
				}
			} else {
				if err != nil {
					t.Errorf("validateID() unexpected error: %v for input: %s", err, tt.tenantID)
				}
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	validTenantID := "550e8400-e29b-41d4-a716-446655440000"
	ctx := context.Background()

	// Create context with tenant
	newCtx, err := NewContext(ctx, validTenantID)
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	// Extract tenant from context
	extractedID, err := FromContext(newCtx)
	if err != nil {
		t.Fatalf("FromContext() failed: %v", err)
	}

	// Verify round trip
	if extractedID != validTenantID {
		t.Errorf("Round trip failed: got %v, want %v", extractedID, validTenantID)
	}

	// Verify HasTenant works
	if !HasTenant(newCtx) {
		t.Errorf("HasTenant() should return true after NewContext()")
	}
}