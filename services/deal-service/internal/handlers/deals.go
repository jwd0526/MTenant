package handlers

import (
	"crm-platform/deal-service/database"
	"crm-platform/deal-service/internal/db"
	"crm-platform/deal-service/internal/errors"
	"crm-platform/deal-service/internal/models"
	"crm-platform/deal-service/tenant"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// HANDLER STRUCT

// Deal handler with tenant-aware database pool
type DealHandler struct {
	tenantPool *tenant.TenantPool
}

// Create new deal handler with tenant-aware database dependencies
func NewDealHandler(pool *database.Pool) *DealHandler {
	return &DealHandler{
		tenantPool: tenant.NewTenantPool(pool),
	}
}

// Create new deal handler with existing tenant pool (for testing)
func NewDealHandlerWithTenantPool(tenantPool *tenant.TenantPool) *DealHandler {
	return &DealHandler{
		tenantPool: tenantPool,
	}
}

// CORE HANDLERS

// Create new deal with validation and automatic tenant isolation
func (h *DealHandler) CreateDeal(c *gin.Context) {
	// Parse and validate request JSON
	var req models.CreateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("failed to validate request JSON").Error()})
		return
	}

	// Add user context data (created_by, owner_id)
	id := extractUserID(c)
	if id == "" {
		return
	}

	// Convert request to SQLC params
	params := h.convertToCreateParams(req, id)

	// Execute database operation with automatic tenant isolation
	queries := db.New(h.tenantPool)
	deal, err := queries.CreateDeal(c.Request.Context(), params)
	if err != nil {
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to create deal").Error()})
		return
	}

	// Convert result to response model
	response := h.convertToResponse(deal)

	// Return JSON response
	c.JSON(201, response)
}

// Get single deal by ID with related data and automatic tenant isolation
func (h *DealHandler) GetDeal(c *gin.Context) {
	// 1. Extract and validate deal ID from URL params
	dealIDStr := c.Param("id")
	dealID, err := strconv.Atoi(dealIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("invalid deal ID").Error()})
		return
	}

	// 2. Query deal with related data using SQLC and automatic tenant isolation
	queries := db.New(h.tenantPool)
	deal, err := queries.GetDealByID(c.Request.Context(), int32(dealID))
	if err != nil {
		// Check for both sql.ErrNoRows and pgx.ErrNoRows
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": errors.ErrDeal("deal not found").Error()})
			return
		}
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to get deal").Error()})
		return
	}

	// 3. Convert to response with calculated fields
	response := h.convertToResponse(deal)

	// 4. Return JSON response
	c.JSON(200, response)
}

// Update existing deal with partial data and automatic tenant isolation
func (h *DealHandler) UpdateDeal(c *gin.Context) {
	// 1. Extract and validate deal ID from URL params
	dealIDStr := c.Param("id")
	dealID, err := strconv.Atoi(dealIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("invalid deal ID").Error()})
		return
	}

	// 2. Parse update request (partial fields)
	var req models.UpdateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("invalid update request").Error()})
		return
	}

	// 3. Add user context data (updated_by)
	userID := extractUserID(c)
	if userID == "" {
		return
	}

	// 4. Convert to SQLC update params
	params := h.convertToUpdateParams(int32(dealID), req, userID)

	// 5. Execute update operation with automatic tenant isolation
	queries := db.New(h.tenantPool)
	deal, err := queries.UpdateDeal(c.Request.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": errors.ErrDeal("deal not found").Error()})
			return
		}
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to update deal").Error()})
		return
	}

	// 6. Return updated deal response
	response := h.convertToResponse(deal)
	c.JSON(200, response)
}

// List deals with pagination and filtering with automatic tenant isolation
func (h *DealHandler) ListDeals(c *gin.Context) {
	// 1. Parse query parameters for pagination/filters
	var query models.ListDealsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		// Use defaults if parsing fails
		query.Page = 1
		query.Limit = 20
	}

	// 2. Set default pagination values
	offset, limit := calculatePagination(query.Page, query.Limit)

	// 3. Execute paginated query with automatic tenant isolation
	queries := db.New(h.tenantPool)
	deals, err := queries.ListDeals(c.Request.Context(), db.ListDealsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to list deals").Error()})
		return
	}

	// 4. Convert deals to response format and apply filters
	var dealResponses []models.DealResponse
	for _, deal := range deals {
		dealResponse := h.convertToResponse(deal)
		
		// Apply stage filter if specified
		if query.Stage != nil && *query.Stage != "" {
			if dealResponse.Stage != *query.Stage {
				continue // Skip deals that don't match the stage filter
			}
		}
		
		// Apply owner filter if specified
		if query.OwnerID != nil {
			if dealResponse.OwnerID == nil || *dealResponse.OwnerID != *query.OwnerID {
				continue // Skip deals that don't match the owner filter
			}
		}
		
		// Apply company filter if specified
		if query.CompanyID != nil {
			if dealResponse.CompanyID == nil || *dealResponse.CompanyID != *query.CompanyID {
				continue // Skip deals that don't match the company filter
			}
		}
		
		// Apply date range filters if specified
		if query.ExpectedCloseFrom != nil && dealResponse.ExpectedCloseDate != nil {
			if dealResponse.ExpectedCloseDate.Before(*query.ExpectedCloseFrom) {
				continue // Skip deals before the from date
			}
		}
		
		if query.ExpectedCloseTo != nil && dealResponse.ExpectedCloseDate != nil {
			if dealResponse.ExpectedCloseDate.After(*query.ExpectedCloseTo) {
				continue // Skip deals after the to date
			}
		}
		
		dealResponses = append(dealResponses, dealResponse)
	}

	// 5. Get total count for proper pagination
	totalCount, err := queries.CountDeals(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to count deals").Error()})
		return
	}

	// Calculate proper pagination metadata
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	// 6. Return paginated response
	response := models.DealListResponse{
		Deals: dealResponses,
		Pagination: models.PaginationMeta{
			Page:       query.Page,
			Limit:      int(limit),
			TotalCount: int(totalCount),
			TotalPages: totalPages,
		},
	}

	c.JSON(200, response)
}

// Get pipeline view with stage analytics and automatic tenant isolation
func (h *DealHandler) GetPipelineView(c *gin.Context) {
	// 1. Query deals by stage analytics with automatic tenant isolation
	queries := db.New(h.tenantPool)
	stageData, err := queries.GetDealsByStage(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to get pipeline data: " + err.Error()).Error()})
		return
	}

	// 2. Convert to response format
	var stages []models.PipelineStage
	var totals models.PipelineTotals

	for _, stage := range stageData {
		pipelineStage := models.PipelineStage{
			Stage:         stage.Stage,
			DealCount:     int(stage.DealCount),
			TotalValue:    h.convertInterfaceToFloat64Safe(stage.TotalValue),
			WeightedValue: h.convertInterfaceToFloat64Safe(stage.WeightedValue),
			Deals:         []models.DealResponse{}, // Simplified - would need separate query for deals per stage
		}
		stages = append(stages, pipelineStage)

		// Add to totals
		totals.TotalDeals += int(stage.DealCount)
		totals.TotalValue += h.convertInterfaceToFloat64Safe(stage.TotalValue)
		totals.TotalWeightedValue += h.convertInterfaceToFloat64Safe(stage.WeightedValue)
	}

	// 3. Return pipeline response
	response := models.PipelineViewResponse{
		Stages: stages,
		Totals: totals,
	}

	c.JSON(200, response)
}

// Get deals filtered by owner ID with automatic tenant isolation
func (h *DealHandler) GetDealsByOwner(c *gin.Context) {
	// 1. Extract and validate owner ID from URL params
	ownerIDStr := c.Param("id")
	ownerID, err := strconv.Atoi(ownerIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("invalid owner ID").Error()})
		return
	}

	// 2. Query deals by owner with automatic tenant isolation
	queries := db.New(h.tenantPool)
	ownerID32 := int32(ownerID)
	deals, err := queries.GetDealsByOwner(c.Request.Context(), &ownerID32)
	if err != nil {
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to get deals by owner").Error()})
		return
	}

	// 3. Convert deals to response format
	var dealResponses []models.DealResponse
	for _, deal := range deals {
		dealResponses = append(dealResponses, h.convertDealToResponse(deal))
	}

	// 4. Return deals array (no pagination wrapper for owner endpoint)
	c.JSON(200, dealResponses)
}

// Close deal with final stage and close date with automatic tenant isolation
func (h *DealHandler) CloseDeal(c *gin.Context) {
	// 1. Extract and validate deal ID from URL params
	dealIDStr := c.Param("id")
	dealID, err := strconv.Atoi(dealIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("invalid deal ID").Error()})
		return
	}

	// 2. Parse close deal request
	var req models.CloseDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("invalid close request").Error()})
		return
	}

	// 3. Add user context data (updated_by)
	userID := extractUserID(c)
	if userID == "" {
		return
	}

	// 4. Set actual close date if not provided
	closeDate := req.ActualCloseDate
	if closeDate == nil {
		now := time.Now()
		closeDate = &now
	}

	// 5. Execute close deal operation with automatic tenant isolation
	queries := db.New(h.tenantPool)
	deal, err := queries.CloseDeal(c.Request.Context(), db.CloseDealParams{
		ID:              int32(dealID),
		Stage:           req.Stage,
		ActualCloseDate: h.convertTimeToNullTime(closeDate),
	})
	if err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			c.JSON(404, gin.H{"error": errors.ErrDeal("deal not found").Error()})
			return
		}
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to close deal").Error()})
		return
	}

	// 6. Return closed deal response
	response := h.convertDealToResponse(deal)
	c.JSON(200, response)
}

// Delete existing deal with automatic tenant isolation
func (h *DealHandler) DeleteDeal(c *gin.Context) {
	// 1. Extract and validate deal ID from URL params
	dealIDStr := c.Param("id")
	dealID, err := strconv.Atoi(dealIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": errors.ErrValidation("invalid deal ID").Error()})
		return
	}

	// 2. Execute delete operation with automatic tenant isolation
	queries := db.New(h.tenantPool)
	rowsAffected, err := queries.DeleteDeal(c.Request.Context(), int32(dealID))
	if err != nil {
		c.JSON(500, gin.H{"error": errors.ErrDatabase("failed to delete deal").Error()})
		return
	}
	
	// Check if any rows were affected (deal found and deleted)
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": errors.ErrDeal("deal not found").Error()})
		return
	}

	// 3. Return success response (204 No Content)
	c.Status(204)
}

// HELPERS

// ensureTenantIsolation is no longer needed - TenantPool handles isolation automatically

// Convert create request to SQLC parameters with user context
func (h *DealHandler) convertToCreateParams(req models.CreateDealRequest, userID string) db.CreateDealParams {
	// Map request fields to SQLC CreateDealParams
	dbReq := db.CreateDealParams{
		Title:             req.Title,
		Value:             h.convertFloat64ToNumeric(req.Value),
		Probability:       h.convertFloat64ToInt32Ptr(req.Probability),
		Stage:             req.Stage,
		PrimaryContactID:  req.PrimaryContactID,
		CompanyID:         req.CompanyID,
		OwnerID:           h.convertStringToInt32Ptr(userID),
		ExpectedCloseDate: h.convertTimeToNullTime(req.ExpectedCloseDate),
		Source:            req.DealSource,
		Description:       req.Description,
		CreatedBy:         h.convertStringToInt32Ptr(userID),
	}
	
	return dbReq
}

// Convert update request to SQLC parameters with user context
func (h *DealHandler) convertToUpdateParams(dealID int32, req models.UpdateDealRequest, _ string) db.UpdateDealParams {
	// Map partial update fields to SQLC UpdateDealParams
	dbReq := db.UpdateDealParams{
		ID: dealID,
		Title: func() string {
			if req.Title != nil {
				return *req.Title
			}
			return "" // Will need to handle this differently - get current value
		}(),
		Value:             h.convertFloat64ToNumeric(req.Value),
		Probability:       h.convertFloat64ToInt32Ptr(req.Probability),
		Stage: func() string {
			if req.Stage != nil {
				return *req.Stage
			}
			return ""
		}(),
		PrimaryContactID:  req.PrimaryContactID,
		CompanyID:         req.CompanyID,
		ExpectedCloseDate: h.convertTimeToNullTime(req.ExpectedCloseDate),
		Source:            req.DealSource,
		Description:       req.Description,
	}
	
	return dbReq
}

// Convert SQLC deal result to response model with calculated fields
func (h *DealHandler) convertToResponse(deal interface{}) models.DealResponse {
	// Handle different SQLC result types
	switch d := deal.(type) {
	case db.Deal:
		return h.convertDealToResponse(d)
	case db.GetDealByIDRow:
		return h.convertGetDealByIDRowToResponse(d)
	case db.ListDealsRow:
		return h.convertListDealsRowToResponse(d)
	default:
		// Return empty response for unsupported types
		return models.DealResponse{}
	}
}

// Convert basic Deal struct to response
func (h *DealHandler) convertDealToResponse(deal db.Deal) models.DealResponse {
	return models.DealResponse{
		ID:                deal.ID,
		Title:             deal.Title,
		Value:             h.convertNumericToFloat64(deal.Value),
		Probability:       h.convertInt32PtrToFloat64(deal.Probability),
		Stage:             deal.Stage,
		PrimaryContactID:  deal.PrimaryContactID,
		CompanyID:         deal.CompanyID,
		OwnerID:           deal.OwnerID,
		ExpectedCloseDate: h.convertNullTimeToTime(deal.ExpectedCloseDate),
		ActualCloseDate:   h.convertNullTimeToTime(deal.ActualCloseDate),
		DealSource:        deal.Source,
		Description:       deal.Description,
		CreatedAt:         deal.CreatedAt,
		UpdatedAt:         deal.UpdatedAt,
		CreatedBy:         deal.CreatedBy,
		// Calculated fields
		DealAge:           h.calculateDealAge(deal.CreatedAt),
		DaysUntilClose:    h.calculateDaysUntilClose(deal.ExpectedCloseDate),
		WeightedValue:     h.calculateWeightedValueInt32(deal.Value, deal.Probability),
	}
}

// Convert GetDealByIDRow to response (includes related names)
func (h *DealHandler) convertGetDealByIDRowToResponse(deal db.GetDealByIDRow) models.DealResponse {
	return models.DealResponse{
		ID:                deal.ID,
		Title:             deal.Title,
		Value:             h.convertNumericToFloat64(deal.Value),
		Probability:       h.convertInt32PtrToFloat64(deal.Probability),
		Stage:             deal.Stage,
		PrimaryContactID:  deal.PrimaryContactID,
		CompanyID:         deal.CompanyID,
		OwnerID:           deal.OwnerID,
		ExpectedCloseDate: h.convertNullTimeToTime(deal.ExpectedCloseDate),
		ActualCloseDate:   h.convertNullTimeToTime(deal.ActualCloseDate),
		DealSource:        deal.Source,
		Description:       deal.Description,
		CreatedAt:         deal.CreatedAt,
		UpdatedAt:         deal.UpdatedAt,
		CreatedBy:         deal.CreatedBy,
		// Related names from joins
		PrimaryContactName: h.convertInterfaceToString(deal.PrimaryContactName),
		CompanyName:        deal.CompanyName,
		OwnerName:          h.convertInterfaceToString(deal.OwnerName),
		// Calculated fields
		DealAge:           h.calculateDealAge(deal.CreatedAt),
		DaysUntilClose:    h.calculateDaysUntilClose(deal.ExpectedCloseDate),
		WeightedValue:     h.calculateWeightedValueInt32(deal.Value, deal.Probability),
	}
}

// Convert ListDealsRow to response (includes related names from joins)
func (h *DealHandler) convertListDealsRowToResponse(deal db.ListDealsRow) models.DealResponse {
	return models.DealResponse{
		ID:                deal.ID,
		Title:             deal.Title,
		Value:             h.convertNumericToFloat64(deal.Value),
		Probability:       h.convertInt32PtrToFloat64(deal.Probability),
		Stage:             deal.Stage,
		PrimaryContactID:  deal.PrimaryContactID,
		CompanyID:         deal.CompanyID,
		OwnerID:           deal.OwnerID,
		ExpectedCloseDate: h.convertNullTimeToTime(deal.ExpectedCloseDate),
		ActualCloseDate:   h.convertNullTimeToTime(deal.ActualCloseDate),
		DealSource:        deal.Source,
		Description:       deal.Description,
		CreatedAt:         deal.CreatedAt,
		UpdatedAt:         deal.UpdatedAt,
		CreatedBy:         deal.CreatedBy,
		// Related names from joins
		PrimaryContactName: h.convertInterfaceToString(deal.PrimaryContactName),
		CompanyName:        deal.CompanyName,
		OwnerName:          h.convertInterfaceToString(deal.OwnerName),
		// Calculated fields
		DealAge:           h.calculateDealAge(deal.CreatedAt),
		DaysUntilClose:    h.calculateDaysUntilClose(deal.ExpectedCloseDate),
		WeightedValue:     h.calculateWeightedValueInt32(deal.Value, deal.Probability),
	}
}

// CONVERSION HELPERS

// Convert *float64 to pgtype.Numeric
func (h *DealHandler) convertFloat64ToNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	var n pgtype.Numeric
	err := n.Scan(fmt.Sprintf("%f", *f))
	if err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}

// Convert *time.Time to sql.NullTime
func (h *DealHandler) convertTimeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// Convert string to *int32 (for user IDs)
func (h *DealHandler) convertStringToInt32Ptr(s string) *int32 {
	if id, err := strconv.Atoi(s); err == nil {
		id32 := int32(id)
		return &id32
	}
	return nil
}

// Convert *float64 to *int32 (for probability)
func (h *DealHandler) convertFloat64ToInt32Ptr(f *float64) *int32 {
	if f == nil {
		return nil
	}
	val := int32(*f)
	return &val
}

// Convert *int32 to *float64 (for probability in response)
func (h *DealHandler) convertInt32PtrToFloat64(i *int32) *float64 {
	if i == nil {
		return nil
	}
	val := float64(*i)
	return &val
}

// Convert pgtype.Numeric to *float64
func (h *DealHandler) convertNumericToFloat64(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	
	// Use Float64() method to get the float64 value
	f, err := n.Float64Value()
	if err != nil {
		return nil
	}
	return &f.Float64
}

// Convert sql.NullTime to *time.Time
func (h *DealHandler) convertNullTimeToTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// Convert interface{} to *string (for SQLC joined fields)
func (h *DealHandler) convertInterfaceToString(i interface{}) *string {
	if i == nil {
		return nil
	}
	
	if s, ok := i.(string); ok {
		return &s
	}
	return nil
}

// Convert interface{} to float64 (safe version for SQLC aggregations)
func (h *DealHandler) convertInterfaceToFloat64Safe(i interface{}) float64 {
	if i == nil {
		return 0.0
	}
	
	switch v := i.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case pgtype.Numeric:
		if !v.Valid {
			return 0.0
		}
		var f float64
		if err := v.Scan(&f); err != nil {
			return 0.0
		}
		return f
	default:
		return 0.0
	}
}

// CALCULATION HELPERS

// Calculate deal age in days
func (h *DealHandler) calculateDealAge(createdAt time.Time) int {
	return int(time.Since(createdAt).Hours() / 24)
}

// Calculate days until expected close
func (h *DealHandler) calculateDaysUntilClose(expectedClose sql.NullTime) *int {
	if !expectedClose.Valid {
		return nil
	}
	
	days := int(time.Until(expectedClose.Time).Hours() / 24)
	return &days
}

// Calculate weighted value (value * probability / 100) - for int32 probability
func (h *DealHandler) calculateWeightedValueInt32(value pgtype.Numeric, probability *int32) *float64 {
	if !value.Valid || probability == nil {
		return nil
	}
	
	// Get float64 value from numeric
	f, err := value.Float64Value()
	if err != nil {
		return nil
	}
	
	p := float64(*probability)
	weighted := f.Float64 * (p / 100)
	return &weighted
}


// Extract user ID from Gin context (set by auth middleware)
func extractUserID(c *gin.Context) string {
	id := c.GetString("user_id")
	if id == "" {
		c.JSON(400, gin.H{"error": errors.ErrHandler("could not extract user id").Error()})
		return ""
	}
	return id
}

// Convert pagination query to offset/limit for SQLC
func calculatePagination(page, limit int) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	return int32(offset), int32(limit)
}