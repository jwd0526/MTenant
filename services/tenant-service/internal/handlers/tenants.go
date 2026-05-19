package handlers

import (
	"net/http"
	"strings"

	"crm-platform/tenant-service/internal/models"
	"crm-platform/tenant-service/internal/services"

	"github.com/gin-gonic/gin"
)

// TenantHandler handles HTTP requests for tenant operations
type TenantHandler struct {
	tenantService *services.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// CreateTenant handles POST /internal/tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req models.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request format: " + err.Error(),
		})
		return
	}

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// GetTenant handles GET /internal/tenants/:id
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Tenant ID required",
		})
		return
	}

	tenant, err := h.tenantService.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// GetTenantBySubdomain handles GET /internal/tenants/subdomain/:subdomain
func (h *TenantHandler) GetTenantBySubdomain(c *gin.Context) {
	subdomain := c.Param("subdomain")
	if subdomain == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Subdomain required",
		})
		return
	}

	tenant, err := h.tenantService.GetTenantBySubdomain(c.Request.Context(), subdomain)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// ListTenants handles GET /internal/tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
	tenants, err := h.tenantService.ListTenants(c.Request.Context())
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenants)
}

// UpdateTenant handles PUT /internal/tenants/:id
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Tenant ID required",
		})
		return
	}

	var req models.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request format: " + err.Error(),
		})
		return
	}

	tenant, err := h.tenantService.UpdateTenant(c.Request.Context(), tenantID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// GetTenantHealth handles GET /internal/tenants/:id/health
func (h *TenantHandler) GetTenantHealth(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Tenant ID required",
		})
		return
	}

	health, err := h.tenantService.GetTenantHealth(c.Request.Context(), tenantID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	if health.Healthy {
		c.JSON(http.StatusOK, health)
	} else {
		c.JSON(http.StatusServiceUnavailable, health)
	}
}

// CreateTestTenants handles POST /internal/test-tenants
func (h *TenantHandler) CreateTestTenants(c *gin.Context) {
	var req models.BulkCreateTenantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request format: " + err.Error(),
		})
		return
	}

	result, err := h.tenantService.CreateTestTenants(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	status := http.StatusCreated
	if len(result.Failed) > 0 {
		status = http.StatusMultiStatus // 207 for mixed results
	}
	c.JSON(status, result)
}

// DeleteTestTenants handles DELETE /internal/test-tenants
func (h *TenantHandler) DeleteTestTenants(c *gin.Context) {
	if c.Query("confirm") != "true" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Add ?confirm=true to confirm deletion",
		})
		return
	}

	err := h.tenantService.DeleteTestTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test tenants deleted"})
}

// handleServiceError converts service errors to appropriate HTTP responses
func (h *TenantHandler) handleServiceError(c *gin.Context, err error) {
	errMsg := err.Error()

	// Determine HTTP status code based on error type
	switch {
	case strings.Contains(errMsg, "NOT FOUND"):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: errMsg})
	case strings.Contains(errMsg, "DUPLICATE ERROR"):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: errMsg})
	case strings.Contains(errMsg, "VALIDATION ERROR"):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
	case strings.Contains(errMsg, "NOT IMPLEMENTED"):
		c.JSON(http.StatusNotImplemented, models.ErrorResponse{Error: errMsg})
	default:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Internal server error"})
	}
}
