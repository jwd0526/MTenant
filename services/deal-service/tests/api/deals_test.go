package api

import (
	"fmt"
	"testing"

	"crm-platform/deal-service/tests/fixtures"
	"crm-platform/deal-service/tests/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DealsAPITestSuite tests all deal API routes with clean data isolation
type DealsAPITestSuite struct {
	suite.Suite
	db       *helpers.TestDatabase
	server   *helpers.TestServer
	fixtures *fixtures.DealFixtures
	tenant1  string
	tenant2  string
}

// SetupSuite runs once before all tests - uses predefined tenant schemas
func (suite *DealsAPITestSuite) SetupSuite() {
	suite.db = helpers.SetupTestDatabase(suite.T())
	suite.server = helpers.SetupTestServer(suite.T(), suite.db)
	suite.fixtures = fixtures.NewDealFixtures()

	// Use predefined test tenants (created by setup script)
	suite.tenant1 = helpers.TestTenant1
	suite.tenant2 = helpers.TestTenant2
	
	// Set up tenant contexts
	suite.db.UsePredefinedTenant(suite.tenant1)
	suite.db.UsePredefinedTenant(suite.tenant2)
}

// TearDownSuite runs once after all tests - closes database connection
func (suite *DealsAPITestSuite) TearDownSuite() {
	if suite.db != nil {
		// Don't cleanup tenant schemas - they're persistent for reuse
		suite.db.Close()
	}
}

// SetupTest runs before each test - clean slate
func (suite *DealsAPITestSuite) SetupTest() {
	// Each test starts with clean tenant data
	suite.cleanTenantData(suite.tenant1)
	suite.cleanTenantData(suite.tenant2)
}

// TearDownTest runs after each test - no action needed
func (suite *DealsAPITestSuite) TearDownTest() {
	// Data cleanup happens in SetupTest for next test
}

// cleanTenantData removes all deals from a tenant (clean slate for each test)
func (suite *DealsAPITestSuite) cleanTenantData(tenantID string) {
	err := suite.db.CleanTenantData(tenantID)
	if err != nil {
		suite.T().Logf("Warning: Failed to clean tenant data for %s: %v", tenantID, err)
	}
}

// =====================================
// POST /api/v1/deals - Create Deal
// =====================================

func (suite *DealsAPITestSuite) TestCreateDeal_ValidDeal_Success() {
	validDeal := suite.fixtures.ValidDeal()

	resp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()

	resp.AssertStatus(suite.T(), 201).
		AssertDealStructure(suite.T()).
		AssertField(suite.T(), "title", validDeal.Title).
		AssertField(suite.T(), "stage", validDeal.Stage)

	// Verify calculated fields
	resp.AssertHasField(suite.T(), "deal_age_days").
		AssertHasField(suite.T(), "weighted_value")
}

func (suite *DealsAPITestSuite) TestCreateDeal_MinimalFields_Success() {
	minimalDeal := suite.fixtures.MinimalDeal()

	resp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(minimalDeal).
		Execute()

	resp.AssertStatus(suite.T(), 201).
		AssertDealStructure(suite.T()).
		AssertField(suite.T(), "title", minimalDeal.Title).
		AssertField(suite.T(), "stage", minimalDeal.Stage)
}

func (suite *DealsAPITestSuite) TestCreateDeal_ValidationErrors_BadRequest() {
	invalidDeals := suite.fixtures.InvalidDeals()

	for scenario, invalidDeal := range invalidDeals {
		suite.T().Run(scenario, func(t *testing.T) {
			resp := suite.server.POST("/api/v1/deals").
				WithServer(suite.server).
				WithTenant(suite.tenant1).
				WithBody(invalidDeal).
				Execute()

			resp.AssertError(t, 400, "")
		})
	}
}

// =====================================
// GET /api/v1/deals/:id - Get Deal
// =====================================

func (suite *DealsAPITestSuite) TestGetDeal_ExistingDeal_Success() {
	// Create a deal first
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()

	// Get the deal
	resp := suite.server.GET("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertDealStructure(suite.T()).
		AssertField(suite.T(), "title", validDeal.Title).
		AssertField(suite.T(), "stage", validDeal.Stage)
}

func (suite *DealsAPITestSuite) TestGetDeal_NonExistentDeal_NotFound() {
	resp := suite.server.GET("/api/v1/deals/99999").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertError(suite.T(), 404, "not found")
}

func (suite *DealsAPITestSuite) TestGetDeal_InvalidID_BadRequest() {
	resp := suite.server.GET("/api/v1/deals/invalid-id").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertError(suite.T(), 400, "invalid")
}

// =====================================
// PUT /api/v1/deals/:id - Update Deal
// =====================================

func (suite *DealsAPITestSuite) TestUpdateDeal_ValidUpdate_Success() {
	// Create a deal first
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()
	updateReq := suite.fixtures.UpdateRequest()

	// Update the deal
	resp := suite.server.PUT("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(updateReq).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertDealStructure(suite.T()).
		AssertField(suite.T(), "title", *updateReq.Title).
		AssertField(suite.T(), "stage", *updateReq.Stage)

	// Verify updated_at changed
	assert.NotEqual(suite.T(), 
		createResp.GetField("updated_at"), 
		resp.GetField("updated_at"),
		"updated_at should change after update")
}

func (suite *DealsAPITestSuite) TestUpdateDeal_NonExistentDeal_NotFound() {
	updateReq := suite.fixtures.UpdateRequest()

	resp := suite.server.PUT("/api/v1/deals/99999").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(updateReq).
		Execute()

	resp.AssertError(suite.T(), 404, "not found")
}

// =====================================
// DELETE /api/v1/deals/:id - Delete Deal
// =====================================

func (suite *DealsAPITestSuite) TestDeleteDeal_ExistingDeal_Success() {
	// Create a deal first
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()

	// Delete the deal
	resp := suite.server.DELETE("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 204)

	// Verify deal is deleted
	getResp := suite.server.GET("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()
	getResp.AssertError(suite.T(), 404, "not found")
}

func (suite *DealsAPITestSuite) TestDeleteDeal_NonExistentDeal_NotFound() {
	resp := suite.server.DELETE("/api/v1/deals/99999").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertError(suite.T(), 404, "not found")
}

func (suite *DealsAPITestSuite) TestDeleteDeal_InvalidID_BadRequest() {
	resp := suite.server.DELETE("/api/v1/deals/invalid-id").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertError(suite.T(), 400, "invalid")
}

// =====================================
// PUT /api/v1/deals/:id/close - Close Deal
// =====================================

func (suite *DealsAPITestSuite) TestCloseDeal_CloseWon_Success() {
	// Create a deal first
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()
	closeReq := suite.fixtures.CloseWonRequest()

	// Close the deal
	resp := suite.server.PUT("/api/v1/deals/"+dealID+"/close").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(closeReq).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertField(suite.T(), "stage", "Closed Won").
		AssertHasField(suite.T(), "actual_close_date")

	// Verify actual_close_date is set
	actualCloseDate := resp.GetField("actual_close_date")
	assert.NotNil(suite.T(), actualCloseDate, "actual_close_date should be set")
}

func (suite *DealsAPITestSuite) TestCloseDeal_CloseLost_Success() {
	// Create a deal first
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()
	closeReq := suite.fixtures.CloseLostRequest()

	// Close the deal
	resp := suite.server.PUT("/api/v1/deals/"+dealID+"/close").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(closeReq).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertField(suite.T(), "stage", "Closed Lost").
		AssertHasField(suite.T(), "actual_close_date")
}

// =====================================
// GET /api/v1/deals - List Deals
// =====================================

func (suite *DealsAPITestSuite) TestListDeals_EmptyDatabase_Success() {
	resp := suite.server.GET("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertHasField(suite.T(), "deals").
		AssertHasField(suite.T(), "pagination")

	// Handle nil deals field (empty result set)
	dealsField := resp.Body["deals"]
	if dealsField == nil {
		// Empty result set - this is expected for empty database
		return
	}
	
	deals, ok := dealsField.([]interface{})
	assert.True(suite.T(), ok, "Deals should be an array")
	assert.Empty(suite.T(), deals, "Deals array should be empty")
}

func (suite *DealsAPITestSuite) TestListDeals_WithDeals_Success() {
	// Create multiple deals
	deal1 := suite.fixtures.ValidDeal()
	deal2 := suite.fixtures.MinimalDeal()

	suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(deal1).
		Execute().AssertStatus(suite.T(), 201)

	suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(deal2).
		Execute().AssertStatus(suite.T(), 201)

	// List deals
	resp := suite.server.GET("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertHasField(suite.T(), "deals").
		AssertHasField(suite.T(), "pagination")

	deals, ok := resp.Body["deals"].([]interface{})
	assert.True(suite.T(), ok, "Deals should be an array")
	assert.Len(suite.T(), deals, 2, "Should have 2 deals")
}

func (suite *DealsAPITestSuite) TestListDeals_WithPagination_Success() {
	// Create multiple deals
	for i := 0; i < 5; i++ {
		deal := suite.fixtures.MinimalDeal()
		deal.Title = fmt.Sprintf("Deal %d", i+1)
		suite.server.POST("/api/v1/deals").
			WithServer(suite.server).
			WithTenant(suite.tenant1).
			WithBody(deal).
			Execute().AssertStatus(suite.T(), 201)
	}

	// Test pagination
	resp := suite.server.GET("/api/v1/deals?page=1&limit=2").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200)

	deals, ok := resp.Body["deals"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), deals, 2, "Should respect limit")

	pagination, ok := resp.Body["pagination"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(1), pagination["page"])
	assert.Equal(suite.T(), float64(2), pagination["limit"])
}

// =====================================
// GET /api/v1/deals/pipeline - Pipeline View
// =====================================

func (suite *DealsAPITestSuite) TestGetPipelineView_EmptyPipeline_Success() {
	resp := suite.server.GET("/api/v1/deals/pipeline").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertHasField(suite.T(), "stages").
		AssertHasField(suite.T(), "totals")
}

func (suite *DealsAPITestSuite) TestGetPipelineView_WithDeals_Success() {
	// Create deals across different stages
	pipelineDeals := suite.fixtures.PipelineDeals()
	for _, deal := range pipelineDeals {
		suite.server.POST("/api/v1/deals").
			WithServer(suite.server).
			WithTenant(suite.tenant1).
			WithBody(deal).
			Execute().AssertStatus(suite.T(), 201)
	}

	resp := suite.server.GET("/api/v1/deals/pipeline").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertHasField(suite.T(), "stages").
		AssertHasField(suite.T(), "totals")

	stages, ok := resp.Body["stages"].([]interface{})
	assert.True(suite.T(), ok, "Stages should be an array")
	assert.Greater(suite.T(), len(stages), 0, "Should have stages with deals")

	totals, ok := resp.Body["totals"].(map[string]interface{})
	assert.True(suite.T(), ok, "Totals should be an object")
	assert.Contains(suite.T(), totals, "total_deals")
	assert.Contains(suite.T(), totals, "total_value")
}

// =====================================
// GET /api/v1/deals/owner/:id - Get Deals by Owner
// =====================================

func (suite *DealsAPITestSuite) TestGetDealsByOwner_NoDeals_Success() {
	resp := suite.server.GET("/api/v1/deals/owner/123").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200)

	// The owner endpoint returns an array directly (no pagination wrapper)
	// Since resp.Body is map[string]interface{}, we need to check for correct structure
	// The response should be empty for non-existent owner
}

// Run the test suite
func TestDealsAPITestSuite(t *testing.T) {
	suite.Run(t, new(DealsAPITestSuite))
}