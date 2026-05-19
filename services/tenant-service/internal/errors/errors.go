package errors

import "fmt"

// Error functions for tenant service processes
var (
	// Database process errors
	ErrDatabase = func(msg string) error {
		return fmt.Errorf("DATABASE ERROR: %s", msg)
	}

	// Validation process errors
	ErrValidation = func(msg string) error {
		return fmt.Errorf("VALIDATION ERROR: %s", msg)
	}

	// Schema management errors
	ErrSchemaCreation = func(msg string) error {
		return fmt.Errorf("SCHEMA ERROR: %s", msg)
	}

	// Handler process errors
	ErrHandler = func(msg string) error {
		return fmt.Errorf("HANDLER ERROR: %s", msg)
	}

	// Not found errors
	ErrNotFound = func(msg string) error {
		return fmt.Errorf("NOT FOUND: %s", msg)
	}

	// Duplicate resource errors
	ErrDuplicateSubdomain = func() error {
		return fmt.Errorf("DUPLICATE ERROR: subdomain already exists")
	}

	// Not implemented errors
	ErrNotImplemented = func(msg string) error {
		return fmt.Errorf("NOT IMPLEMENTED: %s", msg)
	}
)
