package database

import (
	"sync/atomic"
	"time"
)

// Metrics holds database performance and usage metrics
type Metrics struct {
	// Connection metrics
	TotalConnections    int64
	FailedConnections   int64
	ActiveConnections   int64
	
	// Query metrics
	TotalQueries        int64
	FailedQueries       int64
	QueryDuration       int64 // nanoseconds
	
	// Health check metrics
	HealthChecks        int64
	FailedHealthChecks  int64
	LastHealthCheck     int64 // unix timestamp
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncrementConnections increments the total connections counter
func (m *Metrics) IncrementConnections() {
	atomic.AddInt64(&m.TotalConnections, 1)
}

// IncrementFailedConnections increments the failed connections counter
func (m *Metrics) IncrementFailedConnections() {
	atomic.AddInt64(&m.FailedConnections, 1)
}

// SetActiveConnections sets the current active connections count
func (m *Metrics) SetActiveConnections(count int64) {
	atomic.StoreInt64(&m.ActiveConnections, count)
}

// IncrementQueries increments the total queries counter
func (m *Metrics) IncrementQueries() {
	atomic.AddInt64(&m.TotalQueries, 1)
}

// IncrementFailedQueries increments the failed queries counter
func (m *Metrics) IncrementFailedQueries() {
	atomic.AddInt64(&m.FailedQueries, 1)
}

// AddQueryDuration adds to the total query duration
func (m *Metrics) AddQueryDuration(duration time.Duration) {
	atomic.AddInt64(&m.QueryDuration, duration.Nanoseconds())
}

// IncrementHealthChecks increments the health checks counter
func (m *Metrics) IncrementHealthChecks() {
	atomic.AddInt64(&m.HealthChecks, 1)
}

// IncrementFailedHealthChecks increments the failed health checks counter
func (m *Metrics) IncrementFailedHealthChecks() {
	atomic.AddInt64(&m.FailedHealthChecks, 1)
}

// UpdateLastHealthCheck updates the last health check timestamp
func (m *Metrics) UpdateLastHealthCheck() {
	atomic.StoreInt64(&m.LastHealthCheck, time.Now().Unix())
}

// GetMetrics returns a copy of the current metrics
func (m *Metrics) GetMetrics() Metrics {
	return Metrics{
		TotalConnections:    atomic.LoadInt64(&m.TotalConnections),
		FailedConnections:   atomic.LoadInt64(&m.FailedConnections),
		ActiveConnections:   atomic.LoadInt64(&m.ActiveConnections),
		TotalQueries:        atomic.LoadInt64(&m.TotalQueries),
		FailedQueries:       atomic.LoadInt64(&m.FailedQueries),
		QueryDuration:       atomic.LoadInt64(&m.QueryDuration),
		HealthChecks:        atomic.LoadInt64(&m.HealthChecks),
		FailedHealthChecks:  atomic.LoadInt64(&m.FailedHealthChecks),
		LastHealthCheck:     atomic.LoadInt64(&m.LastHealthCheck),
	}
}