package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"crm-platform/pkg/database"
	"crm-platform/pkg/tenant"
)

// Test tenant setup script - minimal tenant operations needed for deal-service testing
// This simulates what the future tenant-service would do, but only for testing purposes

const (
	// Known test tenant IDs that tests can rely on
	TestTenant1 = "123e4567-e89b-12d3-a456-426614174000"
	TestTenant2 = "456e7890-e89b-12d3-a456-426614174111"  
	TestTenant3 = "789e0123-e89b-12d3-a456-426614174222"
)

func main() {
	action := "setup"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	log.Printf("Running action: %s", action)

	switch action {
	case "setup":
		setupTestTenants()
	case "cleanup":
		cleanupTestTenants()
	case "reset":
		cleanupTestTenants()
		setupTestTenants()
	case "verify":
		verifyTestTenants()
	default:
		fmt.Printf("Usage: %s [setup|cleanup|reset|verify]\n", os.Args[0])
		fmt.Println("  setup   - Create test tenant schemas")
		fmt.Println("  cleanup - Remove test tenant schemas")
		fmt.Println("  reset   - Cleanup then setup (fresh state)")
		fmt.Println("  verify  - Check if test tenant schemas exist")
		os.Exit(1)
	}
}

func setupTestTenants() {
	log.Println("Setting up test tenant schemas for deal-service testing...")

	db := connectToDatabase()
	defer db.Close()

	tenants := []string{TestTenant1, TestTenant2, TestTenant3}

	for i, tenantID := range tenants {
		log.Printf("Creating test tenant %d: %s", i+1, tenantID)
		
		ctx := context.Background()
		schemaName := tenant.GenerateSchemaName(tenantID)
		
		// Check if schema already exists
		exists, err := tenant.SchemaExists(ctx, db, schemaName)
		if err != nil {
			log.Fatalf("Failed to check if schema exists: %v", err)
		}
		
		if exists {
			log.Printf("  Schema %s already exists, skipping", schemaName)
			continue
		}

		// Create schema
		err = tenant.CreateSchema(ctx, db, schemaName)
		if err != nil {
			log.Fatalf("Failed to create schema %s: %v", schemaName, err)
		}

		// Copy template structure
		templateExists, err := tenant.SchemaExists(ctx, db, "tenant_template")
		if err != nil {
			log.Fatalf("Failed to check template schema: %v", err)
		}

		if templateExists {
			err = tenant.CopyTemplateSchema(ctx, db, "tenant_template", schemaName)
			if err != nil {
				log.Fatalf("Failed to copy template to %s: %v", schemaName, err)
			}
			log.Printf("  ✅ Created and populated schema: %s", schemaName)
		} else {
			log.Printf("  ⚠️  Created empty schema (no template): %s", schemaName)
		}

		// Create test seed data
		err = createSeedData(ctx, db, schemaName, tenantID)
		if err != nil {
			log.Printf("  ⚠️  Warning: Failed to create seed data for %s: %v", schemaName, err)
		} else {
			log.Printf("  ✅ Created seed data for testing: %s", schemaName)
		}
	}

	log.Println("✅ Test tenant setup complete!")
}

func cleanupTestTenants() {
	log.Println("Cleaning up test tenant schemas...")

	db := connectToDatabase()
	defer db.Close()

	tenants := []string{TestTenant1, TestTenant2, TestTenant3}

	for i, tenantID := range tenants {
		log.Printf("Removing test tenant %d: %s", i+1, tenantID)
		
		ctx := context.Background()
		schemaName := tenant.GenerateSchemaName(tenantID)
		
		// Drop schema with CASCADE to remove all objects
		_, err := db.Exec(ctx, fmt.Sprintf(`DROP SCHEMA IF EXISTS "%s" CASCADE`, schemaName))
		if err != nil {
			log.Printf("  ⚠️  Warning: Failed to drop schema %s: %v", schemaName, err)
		} else {
			log.Printf("  ✅ Removed schema: %s", schemaName)
		}
	}

	log.Println("✅ Test tenant cleanup complete!")
}

func verifyTestTenants() {
	log.Println("Verifying test tenant schemas...")

	db := connectToDatabase()
	defer db.Close()

	tenants := []string{TestTenant1, TestTenant2, TestTenant3}
	allExist := true

	for i, tenantID := range tenants {
		log.Printf("Checking test tenant %d: %s", i+1, tenantID)
		
		ctx := context.Background()
		schemaName := tenant.GenerateSchemaName(tenantID)
		
		exists, err := tenant.SchemaExists(ctx, db, schemaName)
		if err != nil {
			log.Printf("  ❌ Error checking schema %s: %v", schemaName, err)
			allExist = false
			continue
		}

		if exists {
			// Check if it has the deals table
			var tableCount int
			err = db.QueryRow(ctx, 
				"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = $1 AND table_name = 'deals'",
				schemaName).Scan(&tableCount)
			
			if err != nil {
				log.Printf("  ❌ Error checking deals table in %s: %v", schemaName, err)
				allExist = false
			} else if tableCount == 0 {
				log.Printf("  ⚠️  Schema %s exists but missing deals table", schemaName)
				allExist = false
			} else {
				log.Printf("  ✅ Schema %s exists with deals table", schemaName)
			}
		} else {
			log.Printf("  ❌ Schema %s does not exist", schemaName)
			allExist = false
		}
	}

	if allExist {
		log.Println("✅ All test tenant schemas are ready!")
		os.Exit(0)
	} else {
		log.Println("❌ Some test tenant schemas are missing or incomplete")
		log.Println("Run: go run scripts/setup_test_tenants.go setup")
		os.Exit(1)
	}
}

func connectToDatabase() *database.Pool {
	config, err := database.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify connection
	health := pool.HealthCheck(ctx)
	if !health.Healthy {
		log.Fatalf("Database is not healthy: %s", health.Error)
	}

	return pool
}

// createSeedData inserts test data that the Postman collection expects
func createSeedData(ctx context.Context, pool *database.Pool, schemaName string, _ string) error {
	// Set search path to the tenant schema
	_, err := pool.Exec(ctx, fmt.Sprintf(`SET search_path = "%s"`, schemaName))
	if err != nil {
		return fmt.Errorf("failed to set search path: %v", err)
	}

	// Create test users - using actual schema with 'status' instead of 'active'
	userQueries := []string{
		`INSERT INTO users (id, email, password_hash, first_name, last_name, role, status, email_verified, created_by)
		 VALUES (123, 'john.doe@test.com', '$2a$10$dummy', 'John', 'Doe', 'sales_rep', 'active', true, 1)
		 ON CONFLICT (id) DO NOTHING`,
		
		`INSERT INTO users (id, email, password_hash, first_name, last_name, role, status, email_verified, created_by)
		 VALUES (100, 'sales.manager@test.com', '$2a$10$dummy', 'Sarah', 'Manager', 'manager', 'active', true, 1)
		 ON CONFLICT (id) DO NOTHING`,
	}

	// Create test companies
	companyQueries := []string{
		`INSERT INTO companies (id, name, domain, industry, created_by)
		 VALUES (456, 'Acme Corporation', 'acme.com', 'Technology', 123)
		 ON CONFLICT (id) DO NOTHING`,
		
		`INSERT INTO companies (id, name, domain, industry, created_by)
		 VALUES (789, 'Global Enterprises', 'global.com', 'Consulting', 123)
		 ON CONFLICT (id) DO NOTHING`,
	}

	// Create test contacts
	contactQueries := []string{
		`INSERT INTO contacts (id, first_name, last_name, email, company_id, created_by)
		 VALUES (123, 'Jane', 'Smith', 'jane.smith@acme.com', 456, 123)
		 ON CONFLICT (id) DO NOTHING`,
		
		`INSERT INTO contacts (id, first_name, last_name, email, company_id, created_by)
		 VALUES (789, 'Bob', 'Johnson', 'bob.johnson@global.com', 789, 123)
		 ON CONFLICT (id) DO NOTHING`,
	}

	// Execute all queries
	allQueries := append(userQueries, companyQueries...)
	allQueries = append(allQueries, contactQueries...)

	for _, query := range allQueries {
		_, err := pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to execute seed query: %v", err)
		}
	}

	log.Printf("    Created seed data: 2 users, 2 companies, 2 contacts")
	return nil
}