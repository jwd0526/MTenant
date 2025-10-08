package helpers

// Predefined test tenant ULIDs - these must be created by ../../scripts/setup_test_tenants.go script
const (
	TestTenant1 = "01HK153X003BMPJNJB6JHKXK8T"
	TestTenant2 = "01HK3QGM00Y1FYD4HXDQKHGW4S"  
	TestTenant3 = "01HK69XB00FMWEYR0NBGS5JNS1"
)

// GetTestTenants returns the available test tenant IDs
func GetTestTenants() []string {
	return []string{TestTenant1, TestTenant2, TestTenant3}
}