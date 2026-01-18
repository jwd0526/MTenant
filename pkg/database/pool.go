package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool error definitions
var (
	ErrNilConfig          = fmt.Errorf("config cannot be nil")
	ErrPoolConfigParse    = fmt.Errorf("unable to parse pool config")
	ErrConnectionFailed   = fmt.Errorf("failed to create connection pool")
)

// Pool wraps pgxpool.Pool with additional functionality
type Pool struct {
	*pgxpool.Pool
	config  *Config
	metrics *Metrics
}

// NewPool creates a new database connection pool
func NewPool(ctx context.Context, config *Config) (*Pool, error) {
	// Validate config is not nil
	if config == nil {
		return nil, ErrNilConfig
	}

	// Create pgxpool config from connection string
	pgxConfig, err := pgxpool.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPoolConfigParse, err)
	}

	// Set pool configuration (max connections, timeouts, etc.)
	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime
	pgxConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// Initialize metrics before attempting connection
	metrics := NewMetrics()

	// Create pool with retry logic
	var pool *pgxpool.Pool
	for i := 0; i <= config.MaxRetries; i++ {
		metrics.IncrementConnections()
		pool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
		if err == nil {
			err = pool.Ping(ctx)
			if err != nil {
				pool.Close()
				metrics.IncrementFailedConnections()
			} else {
				break
			}
		} else {
			metrics.IncrementFailedConnections()
		}
		if i < config.MaxRetries { // Don't sleep after last attempt
			log.Printf("Connection attempt %d failed: %v. Retrying in %v...\n",
				i+1, err, config.RetryInterval)
			time.Sleep(config.RetryInterval)
		} else {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%w after %d attempts: %v", ErrConnectionFailed, config.MaxRetries+1, err)
	}

	customPool := &Pool{
		Pool:    pool,           // Embed the pgxpool.Pool
		config:  config,         // Store the config
		metrics: metrics,	 	 // Initialize pool metrics
	}

	// Update active connections count
	customPool.updateActiveConnections()

	return customPool, nil
}

// Close gracefully closes the connection pool
func (p *Pool) Close() {
	// Close the underlying pgxpool
	log.Println("Closing connection pool...")
	p.Pool.Close()
	log.Println("Successfully closed connection pool.")
}

// Stats returns connection pool statistics
func (p *Pool) Stats() *pgxpool.Stat {
	return p.Stat()
}

// GetMetrics returns a copy of the current metrics
func (p *Pool) GetMetrics() Metrics {
	return p.metrics.GetMetrics()
}

// updateActiveConnections updates the active connections metric from pool stats
func (p *Pool) updateActiveConnections() {
	stats := p.Stat()
	p.metrics.SetActiveConnections(int64(stats.AcquiredConns()))
}

// Query wraps pgxpool.Pool.Query with metrics tracking
func (p *Pool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	p.metrics.IncrementQueries()

	rows, err := p.Pool.Query(ctx, sql, args...)
	duration := time.Since(start)
	p.metrics.AddQueryDuration(duration)

	if err != nil {
		p.metrics.IncrementFailedQueries()
		return nil, err
	}

	return rows, nil
}

// QueryRow wraps pgxpool.Pool.QueryRow with metrics tracking
func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	start := time.Now()
	p.metrics.IncrementQueries()

	row := p.Pool.QueryRow(ctx, sql, args...)
	duration := time.Since(start)
	p.metrics.AddQueryDuration(duration)

	// Note: pgx.Row doesn't return errors until Scan() is called
	// We can't track failures here, but we track the query attempt
	return row
}

// Exec wraps pgxpool.Pool.Exec with metrics tracking
func (p *Pool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	p.metrics.IncrementQueries()

	tag, err := p.Pool.Exec(ctx, sql, args...)
	duration := time.Since(start)
	p.metrics.AddQueryDuration(duration)

	if err != nil {
		p.metrics.IncrementFailedQueries()
		return tag, err
	}

	return tag, nil
}