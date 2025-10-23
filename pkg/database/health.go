package database

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Healthy      bool          `json:"healthy"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	Stats        *PoolStats    `json:"stats"`
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	AcquireCount         int64 `json:"acquire_count"`
	AcquireDuration      int64 `json:"acquire_duration_ns"`
	AcquiredConns        int32 `json:"acquired_conns"`
	CanceledAcquireCount int64 `json:"canceled_acquire_count"`
	ConstructingConns    int32 `json:"constructing_conns"`
	EmptyAcquireCount    int64 `json:"empty_acquire_count"`
	IdleConns            int32 `json:"idle_conns"`
	MaxConns             int32 `json:"max_conns"`
	TotalConns           int32 `json:"total_conns"`
}

// HealthCheck performs a health check on the database connection
func (p *Pool) HealthCheck(ctx context.Context) *HealthStatus {
	start := time.Now()

	// Increment health check counter
	p.metrics.IncrementHealthChecks()

	// Create a context with timeout for the health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status := &HealthStatus{
		Stats: p.getPoolStats(),
	}

	// Perform ping test
	if err := p.Pool.Ping(healthCtx); err != nil {
		p.metrics.IncrementFailedHealthChecks()
		status.Healthy = false
		status.Error = fmt.Sprintf("ping failed: %v", err)
		status.ResponseTime = time.Since(start)
		p.metrics.UpdateLastHealthCheck()
		return status
	}

	// Perform simple query test
	var result int
	err := p.Pool.QueryRow(healthCtx, "SELECT 1").Scan(&result)
	if err != nil {
		p.metrics.IncrementFailedHealthChecks()
		status.Healthy = false
		status.Error = fmt.Sprintf("query failed: %v", err)
		status.ResponseTime = time.Since(start)
		p.metrics.UpdateLastHealthCheck()
		return status
	}

	if result != 1 {
		p.metrics.IncrementFailedHealthChecks()
		status.Healthy = false
		status.Error = "unexpected query result"
		status.ResponseTime = time.Since(start)
		p.metrics.UpdateLastHealthCheck()
		return status
	}

	// Update active connections and last health check time
	p.updateActiveConnections()
	p.metrics.UpdateLastHealthCheck()

	status.Healthy = true
	status.ResponseTime = time.Since(start)
	return status
}

// getPoolStats converts pgxpool.Stat to our PoolStats structure
func (p *Pool) getPoolStats() *PoolStats {
	stats := p.Pool.Stat()
	return &PoolStats{
		AcquireCount:         stats.AcquireCount(),
		AcquireDuration:      stats.AcquireDuration().Nanoseconds(),
		AcquiredConns:        stats.AcquiredConns(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:    stats.ConstructingConns(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		IdleConns:            stats.IdleConns(),
		MaxConns:             stats.MaxConns(),
		TotalConns:           stats.TotalConns(),
	}
}

// IsHealthy returns true if the database is healthy
func (p *Pool) IsHealthy(ctx context.Context) bool {
	return p.HealthCheck(ctx).Healthy
}