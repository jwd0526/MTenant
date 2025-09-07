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

// Configuration error definitions
var (
	ErrDatabaseURLRequired = fmt.Errorf("DATABASE_URL environment variable is required")
	ErrInvalidDatabaseURL  = fmt.Errorf("failed to parse DATABASE_URL")
	ErrInvalidPort         = fmt.Errorf("invalid port number")
	ErrDatabaseNameRequired = fmt.Errorf("database name is required in URL")
	ErrPasswordRequired    = fmt.Errorf("password is required in DATABASE_URL")
)

// Config holds database connection configuration
type Config struct {
	Host string						// Database host address
	Port int						// Database port number
	Database string					// Database name
	Username string					// Database username
	Password string					// Database password
	MaxConns int32					// Maximum number of live connections
	MinConns int32					// Minimum number of live connections
	MaxConnLifetime time.Duration 	// Maximum lifetime of a single connection
	MaxConnIdleTime time.Duration	// Maximum idle time before connection closure
	ConnectTimeout time.Duration 	// Timeout for establishing connections
	QueryTimeout time.Duration		// Timeout for individual queries
	MaxRetries int					// Maximum number of connection retry attempts
	RetryInterval time.Duration		// Duration between retry attempts
	SSLMode string 					// SSL mode (disable, prefer, require)
}

// LoadConfigFromEnv loads database configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	// Get database URL from environment
	dbConnString := os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		return nil, ErrDatabaseURLRequired
	}

	// Parse database URL
	u, err := url.Parse(dbConnString)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDatabaseURL, err)
	}

	// Parse host and port from URL.Host
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to split host:port: %w", err)
	}

	// Convert port from string to integer
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPort, err)
	}
	
	// Parse database name from URL path
	_, db, found := strings.Cut(u.Path, "/")
	if !found || db == "" {
		return nil, ErrDatabaseNameRequired
	}
	
	// Parse username and password, ensure password exists
	username := u.User.Username()
	password, hasPassword := u.User.Password()
	if !hasPassword {
		return nil, ErrPasswordRequired
	}

	// Parse SSL mode from URL query parameters
	sslMode := u.Query().Get("sslmode")
	if sslMode == "" {
		// Set default SSL mode based on environment
		env := strings.ToLower(os.Getenv("ENVIRONMENT"))
		if env == "dev" || env == "development" || env == "" {
			sslMode = "disable"
		} else {
			sslMode = "prefer"
		}
	}

	// Create configuration with parsed values and reasonable defaults
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