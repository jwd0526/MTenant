package tenant

import (
    "context"
    "fmt"
    "log"
    "regexp"
    
    "crm-platform/deal-service/database"
    "github.com/jackc/pgx/v5/pgconn"
)

// Executor interface for flexibility with connections, pools, and transactions
type Executor interface {
    Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}

// Schema error definitions
var (
    ErrSchemaNotFound       = fmt.Errorf("schema does not exist")
    ErrInvalidSchema        = fmt.Errorf("schema name invalid")
    ErrBlankSchemaName      = fmt.Errorf("schema name cannot be blank")
    ErrSetSearchPathFailure = fmt.Errorf("failed to set search path")
    ErrFailedSchemaCreation = fmt.Errorf("failed to create schema")
    ErrTemplateNotFound     = fmt.Errorf("template schema does not exist")
    ErrTableCopyFailure     = fmt.Errorf("failed to copy table")
    ErrSeedDataFailure      = fmt.Errorf("failed to copy seed data")
)

// Schema name validation regex - PostgreSQL identifier rules (allow hyphens for UUIDs)
var schemaNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// GenerateSchemaName converts tenant ID to PostgreSQL schema name
func GenerateSchemaName(tenantID string) string {
    return fmt.Sprintf("tenant_%s", tenantID)
}

// ExtractTenantSchema extracts tenant ID from context and generates validated schema name
func ExtractTenantSchema(ctx context.Context) (string, error) {
    tenantID, err := FromContext(ctx)
    if err != nil {
        return "", err
    }
    
    schemaName := GenerateSchemaName(tenantID)
    
    if err := validateSchemaName(schemaName); err != nil {
        return "", err
    }
    
    return schemaName, nil
}

// SetSearchPath sets the PostgreSQL search path for tenant isolation
func SetSearchPath(ctx context.Context, exec Executor, schemaName string) error {
    if err := validateSchemaName(schemaName); err != nil {
        return err
    }

    sql := fmt.Sprintf(`SET search_path TO "%s", public`, schemaName)

    _, err := exec.Exec(ctx, sql)
    if err != nil {
        return fmt.Errorf("%w: %v", ErrSetSearchPathFailure, err)
    }

    return nil
}

// SchemaExists checks if schema exists in PostgreSQL
func SchemaExists(ctx context.Context, pool *database.Pool, schemaName string) (bool, error) {
    if err := validateSchemaName(schemaName); err != nil {
        return false, err
    }

    sql := `SELECT EXISTS(
        SELECT 1 FROM information_schema.schemata 
        WHERE schema_name = $1
    )`

    var exists bool
    err := pool.QueryRow(ctx, sql, schemaName).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check schema existence: %w", err)
    }

    return exists, nil
}

// CreateSchema creates a new PostgreSQL schema
func CreateSchema(ctx context.Context, pool *database.Pool, schemaName string) error {
    if err := validateSchemaName(schemaName); err != nil {
        return err
    }

    sql := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`, schemaName)

    _, err := pool.Exec(ctx, sql)
    if err != nil {
        return fmt.Errorf("%w: %v", ErrFailedSchemaCreation, err)
    }
    
    return nil
}

// CopyTemplateSchema copies table structure from template to new schema
func CopyTemplateSchema(ctx context.Context, pool *database.Pool, templateSchema, targetSchema string) error {
    if err := validateSchemaName(templateSchema); err != nil {
        return fmt.Errorf("invalid template schema: %w", err)
    }

    if err := validateSchemaName(targetSchema); err != nil {
        return fmt.Errorf("invalid target schema: %w", err)
    }

    exists, err := SchemaExists(ctx, pool, templateSchema)
    if err != nil {
        return fmt.Errorf("failed to check template schema: %w", err)
    }
    if !exists {
        return fmt.Errorf("%w: %s", ErrTemplateNotFound, templateSchema)
    }

    tables, err := getTableNames(ctx, pool, templateSchema)
    if err != nil {
        return err
    }

    for _, tableName := range tables {
        if err := copyTable(ctx, pool, templateSchema, targetSchema, tableName); err != nil {
            return err
        }
    }

    if err := copySeedData(ctx, pool, templateSchema, targetSchema); err != nil {
        log.Printf("Warning: %v", err)
    }

    return nil
}

// validateSchemaName validates schema name format
func validateSchemaName(schemaName string) error {
    if schemaName == "" {
        return ErrBlankSchemaName
    }
    
    if !schemaNamePattern.MatchString(schemaName) {
        return ErrInvalidSchema
    }
    
    return nil
}

// getTableNames retrieves all table names from the specified schema
func getTableNames(ctx context.Context, pool *database.Pool, schemaName string) ([]string, error) {
    sql := `SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = $1 
            AND table_type = 'BASE TABLE'
            ORDER BY table_name`

    rows, err := pool.Query(ctx, sql, schemaName)
    if err != nil {
        return nil, fmt.Errorf("failed to get tables from schema %s: %w", schemaName, err)
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var tableName string
        if err := rows.Scan(&tableName); err != nil {
            return nil, fmt.Errorf("failed to scan table name: %w", err)
        }
        tables = append(tables, tableName)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating table names: %w", err)
    }

    return tables, nil
}

// copyTable copies a single table structure from source to target schema
func copyTable(ctx context.Context, pool *database.Pool, sourceSchema, targetSchema, tableName string) error {
    // First, check if table already exists
    existsSQL := fmt.Sprintf(`SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = '%s' AND table_name = '%s'
    )`, targetSchema, tableName)
    
    var exists bool
    err := pool.QueryRow(ctx, existsSQL).Scan(&exists)
    if err != nil {
        return fmt.Errorf("failed to check table existence: %v", err)
    }
    
    if exists {
        log.Printf("Table %s.%s already exists, skipping", targetSchema, tableName)
        return nil
    }
    
    // Use INCLUDING DEFAULTS CONSTRAINTS INDEXES instead of ALL to avoid sequence conflicts
    sql := fmt.Sprintf(`CREATE TABLE "%s"."%s" 
                       (LIKE "%s"."%s" INCLUDING DEFAULTS INCLUDING CONSTRAINTS INCLUDING INDEXES)`, 
                       targetSchema, tableName, sourceSchema, tableName)
    
    _, err = pool.Exec(ctx, sql)
    if err != nil {
        return fmt.Errorf("%w: %s from %s to %s: %v", 
            ErrTableCopyFailure, tableName, sourceSchema, targetSchema, err)
    }
    
    return nil
}

// copySeedData copies initial data for specific tables
func copySeedData(ctx context.Context, pool *database.Pool, sourceSchema, targetSchema string) error {
    seedTables := []string{"roles", "settings", "default_pipeline_stages"}

    for _, tableName := range seedTables {
        // Use ON CONFLICT DO NOTHING to make seed data insertion idempotent
        sql := fmt.Sprintf(`INSERT INTO "%s"."%s" 
                           SELECT * FROM "%s"."%s"
                           ON CONFLICT DO NOTHING`, 
                           targetSchema, tableName, sourceSchema, tableName)
        
        _, err := pool.Exec(ctx, sql)
        if err != nil {
            // Continue with other tables - seed data is optional
            log.Printf("Warning: %v: %s: %v", ErrSeedDataFailure, tableName, err)
            continue
        }
        
        log.Printf("Successfully copied seed data for table: %s", tableName)
    }

    return nil
}