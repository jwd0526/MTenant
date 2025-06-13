package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool with additional functionality
type Pool struct {
	*pgxpool.Pool
	config  *Config
	metrics *Metrics
}

// NewPool creates a new database connection pool
func NewPool(ctx context.Context, config *Config) (*Pool, error) {
	// TODO: Validate config is not nil
	if config == nil {
		return nil, fmt.Errorf("can't parse config, ensure DATABASE_URL is correct.")
	}

	// TODO: Create pgxpool config from connection string
	pgxConfig, err := pgxpool.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	// TODO: Set pool configuration (max connections, timeouts, etc.)
	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime
	pgxConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// TODO: Create pool with retry logic
	var pool *pgxpool.Pool
	for i := 0; i <= config.MaxRetries; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
		if err == nil {
			err = pool.Ping(ctx)
			if err != nil {
				pool.Close()
			} else {
				break
			}
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
		return nil, fmt.Errorf("failed to create connection pool after %d attempts: %w", config.MaxRetries+1, err)
	}
	
	customPool := &Pool{
		Pool:    pool,           // Embed the pgxpool.Pool
		config:  config,         // Store the config
		metrics: NewMetrics(),	 // Init pool metrics
	}

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
	return p.Pool.Stat()
}