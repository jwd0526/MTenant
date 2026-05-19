package services

import (
	"context"
	"fmt"

	"crm-platform/pkg/database"
	"crm-platform/pkg/tenant"
	"crm-platform/tenant-service/internal/db"
	"crm-platform/tenant-service/internal/errors"
	"crm-platform/tenant-service/internal/models"

	"github.com/oklog/ulid/v2"
)

// TenantService handles all tenant business logic
type TenantService struct {
	pool    *database.Pool
	queries *db.Queries
}

// NewTenantService creates a new tenant service instance
func NewTenantService(pool *database.Pool) *TenantService {
	return &TenantService{
		pool:    pool,
		queries: db.New(pool),
	}
}

// CreateTenant creates a new tenant with schema provisioning
func (s *TenantService) CreateTenant(ctx context.Context, req models.CreateTenantRequest) (*models.TenantResponse, error) {
	// Check if subdomain already exists
	exists, err := s.queries.CheckSubdomainExists(ctx, req.Subdomain)
	if err != nil {
		return nil, errors.ErrDatabase(fmt.Sprintf("failed to check subdomain: %v", err))
	}
	if exists {
		return nil, errors.ErrDuplicateSubdomain()
	}

	// Generate new tenant ID (ULID)
	tenantID := ulid.Make().String()

	// Generate schema name
	schemaName := tenant.GenerateSchemaName(tenantID)

	// Create tenant record in database
	result, err := s.queries.CreateTenant(ctx, db.CreateTenantParams{
		ID:         tenantID,
		Name:       req.Name,
		Subdomain:  req.Subdomain,
		SchemaName: schemaName,
	})
	if err != nil {
		return nil, errors.ErrDatabase(fmt.Sprintf("failed to create tenant: %v", err))
	}

	// Create tenant schema
	if err := tenant.CreateSchema(ctx, s.pool, schemaName); err != nil {
		return nil, errors.ErrSchemaCreation(fmt.Sprintf("failed to create schema: %v", err))
	}

	// Copy template schema structure
	templateSchema := "template"
	if err := tenant.CopyTemplateSchema(ctx, s.pool, templateSchema, schemaName); err != nil {
		return nil, errors.ErrSchemaCreation(fmt.Sprintf("failed to copy template: %v", err))
	}

	// Get the full tenant record
	fullTenant, err := s.queries.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrDatabase(fmt.Sprintf("failed to retrieve tenant: %v", err))
	}

	return &models.TenantResponse{
		ID:         fullTenant.ID,
		Name:       fullTenant.Name,
		Subdomain:  fullTenant.Subdomain,
		SchemaName: fullTenant.SchemaName,
		CreatedAt:  result.CreatedAt,
		UpdatedAt:  fullTenant.UpdatedAt,
	}, nil
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(ctx context.Context, tenantID string) (*models.TenantResponse, error) {
	tenant, err := s.queries.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrNotFound(fmt.Sprintf("tenant not found: %v", err))
	}

	return &models.TenantResponse{
		ID:         tenant.ID,
		Name:       tenant.Name,
		Subdomain:  tenant.Subdomain,
		SchemaName: tenant.SchemaName,
		CreatedAt:  tenant.CreatedAt,
		UpdatedAt:  tenant.UpdatedAt,
	}, nil
}

// GetTenantBySubdomain retrieves a tenant by subdomain
func (s *TenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*models.TenantResponse, error) {
	tenant, err := s.queries.GetTenantBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, errors.ErrNotFound(fmt.Sprintf("tenant not found: %v", err))
	}

	return &models.TenantResponse{
		ID:         tenant.ID,
		Name:       tenant.Name,
		Subdomain:  tenant.Subdomain,
		SchemaName: tenant.SchemaName,
		CreatedAt:  tenant.CreatedAt,
		UpdatedAt:  tenant.UpdatedAt,
	}, nil
}

// ListTenants retrieves all tenants
func (s *TenantService) ListTenants(ctx context.Context) ([]models.TenantResponse, error) {
	tenants, err := s.queries.ListAllTenants(ctx)
	if err != nil {
		return nil, errors.ErrDatabase(fmt.Sprintf("failed to list tenants: %v", err))
	}

	result := make([]models.TenantResponse, len(tenants))
	for i, t := range tenants {
		result[i] = models.TenantResponse{
			ID:         t.ID,
			Name:       t.Name,
			Subdomain:  t.Subdomain,
			SchemaName: t.SchemaName,
			CreatedAt:  t.CreatedAt,
			UpdatedAt:  t.CreatedAt, // ListAllTenants doesn't return UpdatedAt
		}
	}

	return result, nil
}

// UpdateTenant updates tenant information
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID string, req models.UpdateTenantRequest) (*models.TenantResponse, error) {
	// Check if tenant exists
	_, err := s.queries.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, errors.ErrNotFound(fmt.Sprintf("tenant not found: %v", err))
	}

	// Update tenant name if provided
	if req.Name != nil {
		err = s.queries.UpdateTenantName(ctx, db.UpdateTenantNameParams{
			ID:   tenantID,
			Name: *req.Name,
		})
		if err != nil {
			return nil, errors.ErrDatabase(fmt.Sprintf("failed to update tenant: %v", err))
		}
	}

	// Get updated tenant
	return s.GetTenant(ctx, tenantID)
}

// GetTenantHealth checks tenant health status
func (s *TenantService) GetTenantHealth(ctx context.Context, tenantID string) (*models.TenantHealthResponse, error) {
	// Get tenant
	tenantRecord, err := s.queries.GetTenantByID(ctx, tenantID)
	if err != nil {
		return &models.TenantHealthResponse{
			TenantID: tenantID,
			Healthy:  false,
			Message:  fmt.Sprintf("tenant not found: %v", err),
		}, nil
	}

	// Check if schema exists
	exists, err := tenant.SchemaExists(ctx, s.pool, tenantRecord.SchemaName)
	if err != nil {
		return &models.TenantHealthResponse{
			TenantID:   tenantID,
			SchemaName: tenantRecord.SchemaName,
			Healthy:    false,
			Message:    fmt.Sprintf("failed to check schema: %v", err),
		}, nil
	}

	if !exists {
		return &models.TenantHealthResponse{
			TenantID:   tenantID,
			SchemaName: tenantRecord.SchemaName,
			Healthy:    false,
			Message:    "schema does not exist",
		}, nil
	}

	return &models.TenantHealthResponse{
		TenantID:   tenantID,
		SchemaName: tenantRecord.SchemaName,
		Healthy:    true,
		Message:    "tenant is healthy",
	}, nil
}

// CreateTestTenants creates multiple test tenants
func (s *TenantService) CreateTestTenants(ctx context.Context, req models.BulkCreateTenantsRequest) (*models.BulkCreateTenantsResponse, error) {
	response := &models.BulkCreateTenantsResponse{
		Created: []models.TenantResponse{},
		Failed:  []models.BulkCreationFailure{},
	}

	for _, tenantReq := range req.Tenants {
		tenant, err := s.CreateTenant(ctx, tenantReq)
		if err != nil {
			response.Failed = append(response.Failed, models.BulkCreationFailure{
				Name:      tenantReq.Name,
				Subdomain: tenantReq.Subdomain,
				Error:     err.Error(),
			})
			continue
		}
		response.Created = append(response.Created, *tenant)
	}

	return response, nil
}

// DeleteTestTenants deletes all test tenants (development only)
func (s *TenantService) DeleteTestTenants(ctx context.Context) error {
	// This is a placeholder - in production you'd want to:
	// 1. List all tenants with test_ prefix
	// 2. Delete their schemas
	// 3. Delete their records
	return errors.ErrNotImplemented("bulk tenant deletion not implemented")
}
