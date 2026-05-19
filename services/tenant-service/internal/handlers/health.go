package handlers

import (
	"crm-platform/pkg/database"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles system health check endpoints
type HealthHandler struct {
	pool *database.Pool
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(pool *database.Pool) *HealthHandler {
	return &HealthHandler{
		pool: pool,
	}
}

// HealthCheck endpoint that returns database health status
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Perform database health check
	health := h.pool.HealthCheck(c.Request.Context())

	// Return appropriate HTTP status
	if health.Healthy {
		c.JSON(200, health)
	} else {
		c.JSON(503, health)
	}
}
