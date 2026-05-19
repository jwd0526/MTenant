package models

// CreateTenantRequest represents a request to create a new tenant
type CreateTenantRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=200"`
	Subdomain string `json:"subdomain" binding:"required,min=3,max=63,alphanum"`
}

// UpdateTenantRequest represents a request to update tenant information
type UpdateTenantRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=200"`
	Subdomain *string `json:"subdomain" binding:"omitempty,min=3,max=63,alphanum"`
}

// CreateInvitationRequest represents a request to invite a user to a tenant
type CreateInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member viewer"`
}

// BulkCreateTenantsRequest represents a request to create multiple test tenants
type BulkCreateTenantsRequest struct {
	Tenants []CreateTenantRequest `json:"tenants" binding:"required,min=1,max=10"`
}

// AcceptInvitationRequest represents a request to accept an invitation
type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}
