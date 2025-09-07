package fixtures

import (
	"strings"
	"time"

	"crm-platform/deal-service/internal/models"
)

// DealFixtures provides test data for deal API testing
type DealFixtures struct{}

// NewDealFixtures creates a new fixtures instance
func NewDealFixtures() *DealFixtures {
	return &DealFixtures{}
}

// ValidDeal returns a complete, valid deal for positive testing
func (f *DealFixtures) ValidDeal() models.CreateDealRequest {
	value := 50000.0
	probability := 75.0
	expectedClose := time.Now().Add(30 * 24 * time.Hour)

	return models.CreateDealRequest{
		Title:             "Enterprise Software License",
		Value:             &value,
		Probability:       &probability,
		Stage:             "Qualified",
		PrimaryContactID:  int32Ptr(123),
		CompanyID:         int32Ptr(456),
		ExpectedCloseDate: &expectedClose,
		DealSource:        stringPtr("Inbound Lead"),
		Description:       stringPtr("Multi-year enterprise software license deal"),
		Notes:             stringPtr("Customer very interested in premium package"),
	}
}

// MinimalDeal returns a deal with only required fields
func (f *DealFixtures) MinimalDeal() models.CreateDealRequest {
	return models.CreateDealRequest{
		Title: "Minimal Deal",
		Stage: "Lead",
	}
}

// HighValueDeal returns a high-value deal for testing calculations
func (f *DealFixtures) HighValueDeal() models.CreateDealRequest {
	value := 1000000.0
	probability := 90.0
	expectedClose := time.Now().Add(60 * 24 * time.Hour)

	return models.CreateDealRequest{
		Title:             "Million Dollar Enterprise Deal",
		Value:             &value,
		Probability:       &probability,
		Stage:             "Negotiation",
		ExpectedCloseDate: &expectedClose,
		Description:       stringPtr("Major enterprise transformation project"),
	}
}

// InvalidDeals returns a map of invalid deals for validation testing
func (f *DealFixtures) InvalidDeals() map[string]models.CreateDealRequest {
	return map[string]models.CreateDealRequest{
		"empty_title": {
			Title: "",
			Stage: "Lead",
		},
		"invalid_stage": {
			Title: "Test Deal",
			Stage: "InvalidStage",
		},
		"negative_value": {
			Title: "Test Deal",
			Stage: "Lead",
			Value: floatPtr(-1000.0),
		},
		"invalid_probability": {
			Title: "Test Deal",
			Stage: "Lead",
			Probability: floatPtr(150.0), // > 100
		},
		"long_title": {
			Title: strings.Repeat("x", 256), // Too long
			Stage: "Lead",
		},
	}
}

// UpdateRequest returns a valid update request
func (f *DealFixtures) UpdateRequest() models.UpdateDealRequest {
	newValue := 75000.0
	newProbability := 85.0
	newStage := "Proposal"

	return models.UpdateDealRequest{
		Title:       stringPtr("Updated Enterprise Deal"),
		Value:       &newValue,
		Probability: &newProbability,
		Stage:       &newStage,
		Description: stringPtr("Updated after client meeting"),
	}
}

// CloseWonRequest returns a request to close a deal as won
func (f *DealFixtures) CloseWonRequest() models.CloseDealRequest {
	closeDate := time.Now()
	return models.CloseDealRequest{
		Stage:           "Closed Won",
		ActualCloseDate: &closeDate,
	}
}

// CloseLostRequest returns a request to close a deal as lost
func (f *DealFixtures) CloseLostRequest() models.CloseDealRequest {
	closeDate := time.Now()
	return models.CloseDealRequest{
		Stage:           "Closed Lost",
		ActualCloseDate: &closeDate,
	}
}

// PipelineDeals returns deals across different pipeline stages
func (f *DealFixtures) PipelineDeals() []models.CreateDealRequest {
	return []models.CreateDealRequest{
		{
			Title:       "Lead Stage Deal",
			Stage:       "Lead",
			Value:       floatPtr(10000.0),
			Probability: floatPtr(25.0),
		},
		{
			Title:       "Qualified Stage Deal",
			Stage:       "Qualified",
			Value:       floatPtr(25000.0),
			Probability: floatPtr(50.0),
		},
		{
			Title:       "Proposal Stage Deal",
			Stage:       "Proposal",
			Value:       floatPtr(40000.0),
			Probability: floatPtr(75.0),
		},
		{
			Title:       "Negotiation Stage Deal",
			Stage:       "Negotiation",
			Value:       floatPtr(60000.0),
			Probability: floatPtr(85.0),
		},
	}
}

// Utility functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func int32Ptr(i int32) *int32 {
	return &i
}