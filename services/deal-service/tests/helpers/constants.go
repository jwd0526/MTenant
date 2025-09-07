package helpers

// Predefined test tenant IDs - these must be created by setup_test_tenants.go script
const (
	TestTenant1 = "123e4567-e89b-12d3-a456-426614174000"
	TestTenant2 = "456e7890-e89b-12d3-a456-426614174111"  
	TestTenant3 = "789e0123-e89b-12d3-a456-426614174222"
)

// GetTestTenants returns the available test tenant IDs
func GetTestTenants() []string {
	return []string{TestTenant1, TestTenant2, TestTenant3}
}