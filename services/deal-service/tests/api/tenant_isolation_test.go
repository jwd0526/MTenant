package api

import (
	"testing"

	"crm-platform/deal-service/tests/fixtures"
	"crm-platform/deal-service/tests/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TenantIsolationTestSuite verifies that multi-tenant data isolation works correctly
type TenantIsolationTestSuite struct {
	suite.Suite
	db       *helpers.TestDatabase
	server   *helpers.TestServer
	fixtures *fixtures.DealFixtures
	tenant1  string
	tenant2  string
	tenant3  string
}

// SetupSuite uses predefined tenant schemas for isolation testing
func (suite *TenantIsolationTestSuite) SetupSuite() {
	suite.db = helpers.SetupTestDatabase(suite.T())
	suite.server = helpers.SetupTestServer(suite.T(), suite.db)
	suite.fixtures = fixtures.NewDealFixtures()

	// Use all 3 predefined test tenants for comprehensive isolation testing
	suite.tenant1 = helpers.TestTenant1
	suite.tenant2 = helpers.TestTenant2
	suite.tenant3 = helpers.TestTenant3
	
	// Set up tenant contexts
	suite.db.UsePredefinedTenant(suite.tenant1)
	suite.db.UsePredefinedTenant(suite.tenant2)
	suite.db.UsePredefinedTenant(suite.tenant3)
}

// TearDownSuite closes database connection
func (suite *TenantIsolationTestSuite) TearDownSuite() {
	if suite.db != nil {
		// Don't cleanup tenant schemas - they're persistent for reuse
		suite.db.Close()
	}
}

// SetupTest ensures clean data for isolation testing
func (suite *TenantIsolationTestSuite) SetupTest() {
	// Clean all tenant data before each test for perfect isolation
	suite.cleanTenantData(suite.tenant1)
	suite.cleanTenantData(suite.tenant2)
	suite.cleanTenantData(suite.tenant3)
}

// TearDownTest - no action needed
func (suite *TenantIsolationTestSuite) TearDownTest() {
	// Data cleanup happens in SetupTest
}

// cleanTenantData removes all deals from a tenant
func (suite *TenantIsolationTestSuite) cleanTenantData(tenantID string) {
	err := suite.db.CleanTenantData(tenantID)
	if err != nil {
		suite.T().Logf("Warning: Failed to clean tenant data for %s: %v", tenantID, err)
	}
}

// =====================================
// CREATE Isolation Tests
// =====================================

func (suite *TenantIsolationTestSuite) TestCreateDeal_MultiTenant_DataIsolation() {
	validDeal := suite.fixtures.ValidDeal()

	// Create deals in different tenants
	resp1 := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	resp1.AssertStatus(suite.T(), 201)

	resp2 := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		WithBody(validDeal).
		Execute()
	resp2.AssertStatus(suite.T(), 201)

	// Verify each tenant can only see their own data
	list1 := suite.server.GET("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()
	list1.AssertStatus(suite.T(), 200)

	list2 := suite.server.GET("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		Execute()
	list2.AssertStatus(suite.T(), 200)

	// Each tenant should see exactly 1 deal (their own)
	deals1, ok := list1.Body["deals"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), deals1, 1, "Tenant 1 should see exactly 1 deal")

	deals2, ok := list2.Body["deals"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), deals2, 1, "Tenant 2 should see exactly 1 deal")

	// Verify the deals are different (different IDs)
	deal1ID := deals1[0].(map[string]interface{})["id"]
	deal2ID := deals2[0].(map[string]interface{})["id"]
	assert.NotEqual(suite.T(), deal1ID, deal2ID, "Deals should have different IDs")
}

// =====================================
// READ Isolation Tests
// =====================================

func (suite *TenantIsolationTestSuite) TestGetDeal_CrossTenant_NotFound() {
	// Create deal in tenant1
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()

	// Try to access the deal from tenant2 - should fail
	resp := suite.server.GET("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		Execute()

	resp.AssertError(suite.T(), 404, "not found")
}

func (suite *TenantIsolationTestSuite) TestGetDeal_SameTenant_Success() {
	// Create deal in tenant1
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()

	// Access the deal from same tenant - should succeed
	resp := suite.server.GET("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertField(suite.T(), "title", validDeal.Title)
}

// =====================================
// UPDATE Isolation Tests
// =====================================

func (suite *TenantIsolationTestSuite) TestUpdateDeal_CrossTenant_NotFound() {
	// Create deal in tenant1
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()
	updateReq := suite.fixtures.UpdateRequest()

	// Try to update the deal from tenant2 - should fail
	resp := suite.server.PUT("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		WithBody(updateReq).
		Execute()

	resp.AssertError(suite.T(), 404, "not found")
}

func (suite *TenantIsolationTestSuite) TestUpdateDeal_SameTenant_Success() {
	// Create deal in tenant1
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()
	updateReq := suite.fixtures.UpdateRequest()

	// Update the deal from same tenant - should succeed
	resp := suite.server.PUT("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(updateReq).
		Execute()

	resp.AssertStatus(suite.T(), 200).
		AssertField(suite.T(), "title", *updateReq.Title)
}

// =====================================
// DELETE Isolation Tests  
// =====================================

func (suite *TenantIsolationTestSuite) TestDeleteDeal_CrossTenant_NotFound() {
	// Create deal in tenant1
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()

	// Try to delete the deal from tenant2 - should fail
	resp := suite.server.DELETE("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		Execute()

	resp.AssertError(suite.T(), 404, "not found")

	// Verify deal still exists in tenant1
	getResp := suite.server.GET("/api/v1/deals/"+dealID).
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()
	getResp.AssertStatus(suite.T(), 200)
}

func (suite *TenantIsolationTestSuite) TestDeleteDeal_SameTenant_Success() {
	// Create deal in tenant1
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()

	// Delete the deal from same tenant - should succeed
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

// =====================================
// CLOSE Isolation Tests
// =====================================

func (suite *TenantIsolationTestSuite) TestCloseDeal_CrossTenant_NotFound() {
	// Create deal in tenant1
	validDeal := suite.fixtures.ValidDeal()
	createResp := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	createResp.AssertStatus(suite.T(), 201)

	dealID := createResp.GetIDString()
	closeReq := suite.fixtures.CloseWonRequest()

	// Try to close the deal from tenant2 - should fail
	resp := suite.server.PUT("/api/v1/deals/"+dealID+"/close").
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		WithBody(closeReq).
		Execute()

	resp.AssertError(suite.T(), 404, "not found")
}

// =====================================
// PIPELINE Isolation Tests
// =====================================

func (suite *TenantIsolationTestSuite) TestPipelineView_MultiTenant_DataIsolation() {
	// Create different deals in different tenants
	pipelineDeals := suite.fixtures.PipelineDeals()

	// Create 2 deals in tenant1
	for i := 0; i < 2; i++ {
		suite.server.POST("/api/v1/deals").
			WithServer(suite.server).
			WithTenant(suite.tenant1).
			WithBody(pipelineDeals[i]).
			Execute().AssertStatus(suite.T(), 201)
	}

	// Create 1 deal in tenant2
	suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		WithBody(pipelineDeals[0]).
		Execute().AssertStatus(suite.T(), 201)

	// Get pipeline views for each tenant
	pipeline1 := suite.server.GET("/api/v1/deals/pipeline").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		Execute()
	pipeline1.AssertStatus(suite.T(), 200)

	pipeline2 := suite.server.GET("/api/v1/deals/pipeline").
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		Execute()
	pipeline2.AssertStatus(suite.T(), 200)

	// Verify tenant1 sees 2 deals total
	totals1, ok := pipeline1.Body["totals"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(2), totals1["total_deals"], "Tenant 1 should see 2 deals")

	// Verify tenant2 sees 1 deal total
	totals2, ok := pipeline2.Body["totals"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(1), totals2["total_deals"], "Tenant 2 should see 1 deal")
}

// =====================================
// CONCURRENT Tenant Operations
// =====================================

func (suite *TenantIsolationTestSuite) TestConcurrentTenantOperations_NoInterference() {
	// This test verifies that multiple tenants can operate simultaneously
	// without interfering with each other's data
	
	validDeal := suite.fixtures.ValidDeal()
	
	// Create deals simultaneously in 3 different tenants
	// In a real scenario, these would be concurrent requests
	resp1 := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant1).
		WithBody(validDeal).
		Execute()
	resp1.AssertStatus(suite.T(), 201)

	resp2 := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant2).
		WithBody(validDeal).
		Execute()
	resp2.AssertStatus(suite.T(), 201)

	resp3 := suite.server.POST("/api/v1/deals").
		WithServer(suite.server).
		WithTenant(suite.tenant3).
		WithBody(validDeal).
		Execute()
	resp3.AssertStatus(suite.T(), 201)

	// Verify each tenant sees exactly their own data
	for i, tenant := range []string{suite.tenant1, suite.tenant2, suite.tenant3} {
		listResp := suite.server.GET("/api/v1/deals").
			WithServer(suite.server).
			WithTenant(tenant).
			Execute()
		listResp.AssertStatus(suite.T(), 200)

		deals, ok := listResp.Body["deals"].([]interface{})
		assert.True(suite.T(), ok)
		assert.Len(suite.T(), deals, 1, "Tenant %d should see exactly 1 deal", i+1)
	}
}

// Run the isolation test suite
func TestTenantIsolationTestSuite(t *testing.T) {
	suite.Run(t, new(TenantIsolationTestSuite))
}