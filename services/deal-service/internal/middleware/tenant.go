package middleware

import (
	"crm-platform/deal-service/internal/errors"
	"crm-platform/deal-service/tenant"

	"github.com/gin-gonic/gin"
)

// Simple tenant middleware, convert gin context to global context
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get tenant ID from Gin context (set in auth middleware)
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			c.JSON(400, gin.H{"error": errors.ErrTenant("tenant context missing").Error()})
			c.Abort()
			return
		}

		// Create go req context with tenant info
		// ID is validated in tenant.NewContext()
		tenantCtx, err := tenant.NewContext(c.Request.Context(), tenantID)
		if err != nil {
			c.JSON(400, gin.H{"error": errors.ErrTenant("invalid tenant context").Error()})
			c.Abort()
			return
		}

		// Set tenant aware context
		c.Request = c.Request.WithContext(tenantCtx)
		c.Next()
	}
}