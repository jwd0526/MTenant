package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"crm-platform/deal-service/internal/handlers"
	"crm-platform/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer manages HTTP testing for all API routes
type TestServer struct {
	Router      *gin.Engine
	DealHandler *handlers.DealHandler
	t           *testing.T
}

// TestRequest represents an HTTP test request
type TestRequest struct {
	Method   string
	URL      string
	Body     interface{}
	Headers  map[string]string
	TenantID string
	UserID   string
}

// TestResponse represents an HTTP test response
type TestResponse struct {
	StatusCode int                    `json:"status_code"`
	Body       map[string]interface{} `json:"body"`
	RawBody    string                 `json:"raw_body"`
	Headers    http.Header            `json:"headers"`
}

// SetupTestServer creates a test HTTP server with all API routes properly configured
func SetupTestServer(t *testing.T, db *TestDatabase) *TestServer {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware in correct order (same as production)
	router.Use(middleware.AuthMiddleware())
	router.Use(middleware.TenantMiddleware())

	// Create deal handler
	dealHandler := handlers.NewDealHandlerWithTenantPool(db.TenantPool)

	// Register ALL API routes (this was the missing piece!)
	v1 := router.Group("/api/v1")
	deals := v1.Group("/deals")
	{
		deals.POST("", dealHandler.CreateDeal)           // POST /api/v1/deals
		deals.GET("", dealHandler.ListDeals)             // GET /api/v1/deals
		deals.GET("/pipeline", dealHandler.GetPipelineView) // GET /api/v1/deals/pipeline
		deals.GET("/owner/:id", dealHandler.GetDealsByOwner) // GET /api/v1/deals/owner/:id
		deals.GET("/:id", dealHandler.GetDeal)           // GET /api/v1/deals/:id
		deals.PUT("/:id", dealHandler.UpdateDeal)        // PUT /api/v1/deals/:id
		deals.PUT("/:id/close", dealHandler.CloseDeal)   // PUT /api/v1/deals/:id/close
		deals.DELETE("/:id", dealHandler.DeleteDeal)     // DELETE /api/v1/deals/:id ← FIX: This was missing!
	}

	return &TestServer{
		Router:      router,
		DealHandler: dealHandler,
		t:           t,
	}
}

// POST creates a POST request builder
func (ts *TestServer) POST(path string) *RequestBuilder {
	return NewRequest(ts.t, "POST", path)
}

// GET creates a GET request builder
func (ts *TestServer) GET(path string) *RequestBuilder {
	return NewRequest(ts.t, "GET", path)
}

// PUT creates a PUT request builder
func (ts *TestServer) PUT(path string) *RequestBuilder {
	return NewRequest(ts.t, "PUT", path)
}

// DELETE creates a DELETE request builder
func (ts *TestServer) DELETE(path string) *RequestBuilder {
	return NewRequest(ts.t, "DELETE", path)
}

// Execute performs the HTTP request and returns response
func (ts *TestServer) Execute(req TestRequest) *TestResponse {
	var bodyReader io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		require.NoError(ts.t, err, "Failed to marshal request body")
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	httpReq := httptest.NewRequest(req.Method, req.URL, bodyReader)

	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set tenant and user headers for development mode
	if req.TenantID == "" {
		req.TenantID = "default-test-tenant"
	}
	if req.UserID == "" {
		req.UserID = "test-user"
	}

	httpReq.Header.Set("X-Tenant-ID", req.TenantID)
	httpReq.Header.Set("X-User-ID", req.UserID)

	recorder := httptest.NewRecorder()
	ts.Router.ServeHTTP(recorder, httpReq)

	var bodyMap map[string]interface{}
	if recorder.Body.Len() > 0 {
		err := json.Unmarshal(recorder.Body.Bytes(), &bodyMap)
		if err != nil {
			ts.t.Logf("Failed to parse JSON response: %v", err)
		}
	}

	return &TestResponse{
		StatusCode: recorder.Code,
		Body:       bodyMap,
		RawBody:    recorder.Body.String(),
		Headers:    recorder.Header(),
	}
}

// RequestBuilder provides a fluent API for building test requests
type RequestBuilder struct {
	req TestRequest
	t   *testing.T
	server *TestServer
}

// NewRequest creates a new request builder
func NewRequest(t *testing.T, method, url string) *RequestBuilder {
	return &RequestBuilder{
		req: TestRequest{
			Method:  method,
			URL:     url,
			Headers: make(map[string]string),
		},
		t: t,
	}
}

// WithServer sets the test server (for fluent execution)
func (rb *RequestBuilder) WithServer(server *TestServer) *RequestBuilder {
	rb.server = server
	return rb
}

// WithBody adds a JSON body to the request
func (rb *RequestBuilder) WithBody(body interface{}) *RequestBuilder {
	rb.req.Body = body
	return rb
}

// WithTenant sets the tenant ID
func (rb *RequestBuilder) WithTenant(tenantID string) *RequestBuilder {
	rb.req.TenantID = tenantID
	return rb
}

// WithUser sets the user ID
func (rb *RequestBuilder) WithUser(userID string) *RequestBuilder {
	rb.req.UserID = userID
	return rb
}

// WithHeader adds a custom header
func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.req.Headers[key] = value
	return rb
}

// Execute performs the request and returns the response
func (rb *RequestBuilder) Execute() *TestResponse {
	require.NotNil(rb.t, rb.server, "Server must be set before executing request")
	return rb.server.Execute(rb.req)
}

// Build returns the constructed TestRequest
func (rb *RequestBuilder) Build() TestRequest {
	return rb.req
}

// RESPONSE ASSERTIONS

// AssertStatus checks the response status code
func (resp *TestResponse) AssertStatus(t *testing.T, expectedStatus int) *TestResponse {
	assert.Equal(t, expectedStatus, resp.StatusCode,
		"Expected status %d, got %d. Response: %s", expectedStatus, resp.StatusCode, resp.RawBody)
	return resp
}

// AssertSuccess validates a successful response (2xx)
func (resp *TestResponse) AssertSuccess(t *testing.T) *TestResponse {
	assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300,
		"Expected success status (2xx), got %d. Response: %s", resp.StatusCode, resp.RawBody)
	return resp
}

// AssertError validates an error response with specific status and message content
func (resp *TestResponse) AssertError(t *testing.T, expectedStatus int, errorContains string) *TestResponse {
	assert.Equal(t, expectedStatus, resp.StatusCode,
		"Expected error status %d, got %d. Response: %s", expectedStatus, resp.StatusCode, resp.RawBody)

	if errorContains != "" {
		errorMsg, exists := resp.Body["error"]
		require.True(t, exists, "Response should contain 'error' field")
		errorStr := fmt.Sprintf("%v", errorMsg)
		assert.Contains(t, strings.ToLower(errorStr), strings.ToLower(errorContains),
			"Error should contain '%s', got: %s", errorContains, errorStr)
	}
	return resp
}

// AssertField checks that a response field has the expected value
func (resp *TestResponse) AssertField(t *testing.T, field string, expectedValue interface{}) *TestResponse {
	actualValue, exists := resp.Body[field]
	assert.True(t, exists, "Response should contain field '%s'", field)
	assert.Equal(t, expectedValue, actualValue, "Field '%s' should equal '%v', got '%v'", field, expectedValue, actualValue)
	return resp
}

// AssertHasField checks that a response contains a specific field
func (resp *TestResponse) AssertHasField(t *testing.T, field string) *TestResponse {
	assert.Contains(t, resp.Body, field, "Response should contain field '%s'", field)
	return resp
}

// AssertDealStructure validates deal response structure
func (resp *TestResponse) AssertDealStructure(t *testing.T) *TestResponse {
	requiredFields := []string{"id", "title", "stage", "created_at", "updated_at"}
	for _, field := range requiredFields {
		resp.AssertHasField(t, field)
	}
	return resp
}

// GetField extracts a field from the response body
func (resp *TestResponse) GetField(field string) interface{} {
	return resp.Body[field]
}

// GetID extracts the ID field as an integer
func (resp *TestResponse) GetID() int {
	id, ok := resp.Body["id"].(float64)
	if !ok {
		return 0
	}
	return int(id)
}

// GetIDString extracts the ID field as a string
func (resp *TestResponse) GetIDString() string {
	id := resp.GetID()
	if id == 0 {
		return ""
	}
	return strconv.Itoa(id)
}