package database

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds database connection configuration
type Config struct {
	Host string						// Adress	
	Port int						// Port no
	Database string					// Database connection string
	Username string					// DB user
	Password string					// DB user password
	MaxConns int32					// Max # live connections
	MinConns int32					// Min # live connections
	MaxConnLifetime time.Duration 	// How long a single connection can exist before being recreated
	MaxConnIdleTime time.Duration	// Idle timeout duration
	ConnectTimeout time.Duration 	// Timeout on connecting to db
	QueryTimeout time.Duration		// Timeout on query
	MaxRetries int					// Max # retries
	RetryInterval time.Duration		// Duration between retries
	SSLMode string 					// SSL mode (disable, prefer, require)
}

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	// Get db url .env
	dbConnString := os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Parse url
	u, err := url.Parse(dbConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	// Parse host and port from URL.Host
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to split host:port: %w", err)
	}

	// Convert port from string literal to int
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %w", err)
	}
	
	// Parse db
	_, db, found := strings.Cut(u.Path, "/")
	if !found || db == "" {
		return nil, fmt.Errorf("database name is required in URL")
	}
	
	// Parse username and pass, ensure password exists
	username := u.User.Username()
	password, hasPassword := u.User.Password()
	if !hasPassword {
		return nil, fmt.Errorf("password is required in DATABASE_URL")
	}

	// Parse ssl mode from query in URL
	sslMode := u.Query().Get("sslmode")
	if sslMode == "" {
		// Set default if empty
		if os.Getenv("ENVIRONMENT") == "dev" {
			sslMode = "disable"
		} else {
			sslMode = "prefer"
		}
	}

	// Create config from parsed variables
	// Set reasonable defaults for connection variables
	config := Config{
		Host:     			host,
		Port:     			portInt,
		Database: 			db,
		Username: 			username,
		Password: 			password,
		MaxConns: 			20,
		MinConns: 			5,
		MaxConnLifetime: 	time.Minute * 60,
		MaxConnIdleTime: 	time.Minute * 5,
		ConnectTimeout: 	time.Second * 30,
		QueryTimeout: 		time.Second * 30,
		MaxRetries: 		5,
		RetryInterval: 		time.Second * 10,
		SSLMode: 			sslMode,
	}

	return &config, nil
}

// ConnectionString returns the connection string for pgx
func (c *Config) ConnectionString() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
        c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}