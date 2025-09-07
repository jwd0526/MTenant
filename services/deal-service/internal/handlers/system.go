package handlers

import (
	"crm-platform/pkg/database"
	"github.com/gin-gonic/gin"
)

// SystemHandler handles system endpoints like health checks
type SystemHandler struct {
	pool *database.Pool
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(pool *database.Pool) *SystemHandler {
	return &SystemHandler{
		pool: pool,
	}
}

// HealthCheck endpoint that returns database health status
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	// Perform database health check
	health := h.pool.HealthCheck(c.Request.Context())
	
	// Return appropriate HTTP status
	if health.Healthy {
		c.JSON(200, health)
	} else {
		c.JSON(503, health)
	}
}