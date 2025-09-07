package models

import (
	"time"
)


// Deal request model
// Omitted for security: OwnerID, TenantID, CreatedBy
type CreateDealRequest struct {
	Title             string         `json:"title" binding:"required,min=1,max=200"`
	Value             *float64		 `json:"value" binding:"omitempty,min=0"`
	Probability       *float64       `json:"probability" binding:"omitempty,min=0,max=100"`
	Stage             string         `json:"stage" binding:"required,oneof=Lead Qualified Proposal Negotiation 'Closed Won' 'Closed Lost'"`
	PrimaryContactID  *int32         `json:"primary_contact_id"`
	CompanyID         *int32         `json:"company_id"`
	ExpectedCloseDate *time.Time     `json:"expected_close_date"`
	DealSource        *string        `json:"deal_source" binding:"omitempty,max=100"`
	Description       *string        `json:"description" binding:"omitempty,max=1000"`
	Notes             *string        `json:"notes" binding:"omitempty,max=2000"`
}

// Update deal model - all fields optional for partial updates
// Omitted for security: OwnerID, TenantID, UpdatedBy
type UpdateDealRequest struct {
	Title             *string    `json:"title" binding:"omitempty,min=1,max=200"`
	Value             *float64   `json:"value" binding:"omitempty,min=0"`
	Probability       *float64   `json:"probability" binding:"omitempty,min=0,max=100"`
	Stage             *string    `json:"stage" binding:"omitempty,oneof=Lead Qualified Proposal Negotiation 'Closed Won' 'Closed Lost'"`
	PrimaryContactID  *int32     `json:"primary_contact_id"`
	CompanyID         *int32     `json:"company_id"`
	ExpectedCloseDate *time.Time `json:"expected_close_date"`
	DealSource        *string    `json:"deal_source" binding:"omitempty,max=100"`
	Description       *string    `json:"description" binding:"omitempty,max=1000"`
	Notes             *string    `json:"notes" binding:"omitempty,max=2000"`
}

// List deals query params
type ListDealsQuery struct {
	// Pagination
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`

	// Filters
	Stage    *string `form:"stage" binding:"omitempty,oneof=Lead Qualified Proposal Negotiation 'Closed Won' 'Closed Lost'"`
	OwnerID  *int32  `form:"owner_id"`
	CompanyID *int32  `form:"company_id"`

	// Date range filters
	ExpectedCloseFrom *time.Time `form:"expected_close_from" time_format:"2006-01-02"`
	ExpectedCloseTo   *time.Time `form:"expected_close_to" time_format:"2006-01-02"`
}

// Move deal between pipeline stages
type MoveDealStageRequest struct {
	Stage string `json:"stage" binding:"required,oneof=Lead Qualified Proposal Negotiation 'Closed Won' 'Closed Lost'"`
}

// Close deal with final stage and close date
type CloseDealRequest struct {
	Stage           string     `json:"stage" binding:"required,oneof='Closed Won' 'Closed Lost'"`
	ActualCloseDate *time.Time `json:"actual_close_date" time_format:"2006-01-02T15:04:05Z07:00"`
}