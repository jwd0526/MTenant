package errors

import "fmt"

// Error functions for each process - reusable with custom messages
var (
	// Authentication process errors
	ErrAuth = func(msg string) error {
		return fmt.Errorf("AUTHENTICATION ERROR: %s", msg)
	}

	// JWT validation process errors  
	ErrJWT = func(msg string) error {
		return fmt.Errorf("JWT ERROR: %s", msg)
	}

	// Tenant isolation process errors
	ErrTenant = func(msg string) error {
		return fmt.Errorf("TENANT ERROR: %s", msg)
	}

	// Permission checking process errors
	ErrPermission = func(msg string) error {
		return fmt.Errorf("PERMISSION ERROR: %s", msg)
	}

	// Database process errors
	ErrDatabase = func(msg string) error {
		return fmt.Errorf("DATABASE ERROR: %s", msg)
	}

	// Validation process errors
	ErrValidation = func(msg string) error {
		return fmt.Errorf("VALIDATION ERROR: %s", msg)
	}

	// Business logic errors
	ErrBusiness = func(msg string) error {
		return fmt.Errorf("BUSINESS ERROR: %s", msg)
	}

	// Handler process errors  
	ErrHandler = func(msg string) error {
		return fmt.Errorf("HANDLER ERROR: %s", msg)
	}

	// Type conversion errors
	ErrConversion = func(msg string) error {
		return fmt.Errorf("CONVERSION ERROR: %s", msg)
	}

	// Service communication errors
	ErrService = func(msg string) error {
		return fmt.Errorf("SERVICE ERROR: %s", msg)
	}
)