package models

import (
	"time"
)

// TenantResponse represents a tenant with all details
type TenantResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Subdomain  string    `json:"subdomain"`
	SchemaName string    `json:"schema_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TenantHealthResponse represents the health status of a tenant
type TenantHealthResponse struct {
	TenantID   string `json:"tenant_id"`
	Healthy    bool   `json:"healthy"`
	SchemaName string `json:"schema_name"`
	Message    string `json:"message,omitempty"`
}

// InvitationResponse represents an invitation with details
type InvitationResponse struct {
	ID         string     `json:"id"`
	TenantID   *string    `json:"tenant_id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// BulkCreateTenantsResponse represents the result of bulk tenant creation
type BulkCreateTenantsResponse struct {
	Created []TenantResponse       `json:"created"`
	Failed  []BulkCreationFailure  `json:"failed,omitempty"`
}

// BulkCreationFailure represents a failed tenant creation
type BulkCreationFailure struct {
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
	Error     string `json:"error"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}
