package models

import (
	"time"
)

// Single deal response with related data and calculated fields
type DealResponse struct {
	ID                int32      `json:"id"`
	Title             string     `json:"title"`
	Value             *float64   `json:"value"`
	Probability       *float64   `json:"probability"`
	Stage             string     `json:"stage"`
	PrimaryContactID  *int32     `json:"primary_contact_id"`
	CompanyID         *int32     `json:"company_id"`
	OwnerID           *int32     `json:"owner_id"`
	ExpectedCloseDate *time.Time `json:"expected_close_date"`
	ActualCloseDate   *time.Time `json:"actual_close_date"`
	DealSource        *string    `json:"deal_source"`
	Description       *string    `json:"description"`
	Notes             *string    `json:"notes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CreatedBy         *int32     `json:"created_by"`
	UpdatedBy         *int32     `json:"updated_by"`
	
	// Related data (from SQLC joins)
	PrimaryContactName *string `json:"primary_contact_name"`
	CompanyName        *string `json:"company_name"`
	OwnerName          *string `json:"owner_name"`
	
	// Calculated fields
	DealAge            int   `json:"deal_age_days"`           // Days since created
	DaysUntilClose     *int  `json:"days_until_close"`       // Days until expected close
	WeightedValue      *float64 `json:"weighted_value"`      // Value * Probability / 100
}

// Paginated deal collection
type DealListResponse struct {
	Deals      []DealResponse  `json:"deals"`
	Pagination PaginationMeta  `json:"pagination"`
}

// Pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// Pipeline view data grouped by stage
type PipelineViewResponse struct {
	Stages []PipelineStage `json:"stages"`
	Totals PipelineTotals  `json:"totals"`
}

// Single pipeline stage with deals
type PipelineStage struct {
	Stage       string         `json:"stage"`
	DealCount   int           `json:"deal_count"`
	TotalValue  float64       `json:"total_value"`
	WeightedValue float64     `json:"weighted_value"`
	Deals       []DealResponse `json:"deals"`
}

// Pipeline totals across all stages
type PipelineTotals struct {
	TotalDeals        int     `json:"total_deals"`
	TotalValue        float64 `json:"total_value"`
	TotalWeightedValue float64 `json:"total_weighted_value"`
}

// Analytics response for reporting endpoints
type AnalyticsResponse struct {
	PipelineAnalytics []StageAnalytics    `json:"pipeline_analytics"`
	MonthlyForecast   []MonthlyForecast   `json:"monthly_forecast"`
	SalesRepPerformance []SalesRepPerformance `json:"sales_rep_performance"`
}

// Analytics data per stage
type StageAnalytics struct {
	Stage         string  `json:"stage"`
	DealCount     int64   `json:"deal_count"`
	TotalValue    float64 `json:"total_value"`
	WeightedValue float64 `json:"weighted_value"`
}

// Monthly revenue forecast
type MonthlyForecast struct {
	Month           time.Time `json:"month"`
	ExpectedRevenue float64   `json:"expected_revenue"`
	DealCount       int64     `json:"deal_count"`
}

// Sales rep performance metrics
type SalesRepPerformance struct {
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	DealsClosed  int64   `json:"deals_closed"`
	TotalRevenue float64 `json:"total_revenue"`
}

// Standard error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// Error detail structure
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code"`
}