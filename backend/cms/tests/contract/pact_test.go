package contract

import (
	"fmt"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
)

// Test data
var commonHeaders = map[string]string{
	"Content-Type": "application/json",
}

// pactURL determines where pacts will be saved
var pactURL = fmt.Sprintf("%s/pacts", os.Getenv("PACT_DIR"))

// Set up Pact client
func setupPact() *dsl.Pact {
	return &dsl.Pact{
		Consumer: "cms-service",
		Provider: "content-delivery-service",
		LogDir:   "logs",
		PactDir:  "./pacts",
		LogLevel: "INFO",
	}
}

func TestPactCmsToContentDelivery(t *testing.T) {
	// Initialize Pact
	pact := setupPact()
	defer pact.Teardown()

	// Setup expected interactions
	pact.
		AddInteraction().
		Given("Content exists for course 123").
		UponReceiving("A request for content for course 123").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.Term("/api/contents/course/123", "/api/contents/course/[0-9]+"),
			Headers: map[string]dsl.Matcher{
				"Authorization": dsl.Term("Bearer token", "Bearer [a-zA-Z0-9-_.]+"),
				"Content-Type":  dsl.String("application/json"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Headers: map[string]dsl.Matcher{
				"Content-Type": dsl.String("application/json; charset=utf-8"),
			},
			Body: dsl.Match(map[string]interface{}{
				"courseID": dsl.String("123"),
				"contents": dsl.EachLike(map[string]interface{}{
					"id":          dsl.Like("abc123"),
					"title":       dsl.Like("Introduction to Course"),
					"type":        dsl.Term("video", "video|document|quiz"),
					"url":         dsl.Term("https://example.com/content/video1.mp4", "https://.+"),
					"description": dsl.Like("Course introduction video"),
				}, 1),
			}),
		})

	// Verify - Run the actual test
	if err := pact.Verify(func() error {
		// Make the actual API call to the provider
		// This would be your service client code that makes HTTP requests
		// For testing purposes, we're using http.DefaultClient and manually constructing requests

		// Instead of making a real call here, we're just validating the contract definition
		// In a real scenario, you would make API calls to the mock server set up by pact-go
		// Example:
		//   resp, err := http.Get(fmt.Sprintf("%s/api/contents/course/123", pact.Server.URL))
		//   if err != nil || resp.StatusCode != 200 {
		//     return fmt.Errorf("Failed to make API call")
		//   }

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// Write pact files to disk
	if err := pact.WritePact(); err != nil {
		t.Fatal(err)
	}
}

func TestPactContentDeliveryStatusEndpoint(t *testing.T) {
	// Initialize Pact
	pact := setupPact()
	defer pact.Teardown()

	// Test the health/status endpoint
	pact.
		AddInteraction().
		Given("Content delivery service is healthy").
		UponReceiving("A health check request").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   dsl.String("/health"),
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Headers: map[string]dsl.Matcher{
				"Content-Type": dsl.String("application/json; charset=utf-8"),
			},
			Body: dsl.Match(map[string]interface{}{
				"status":  dsl.String("ok"),
				"service": dsl.Like("content-delivery"),
				"version": dsl.Term("1.0.0", "[0-9]+\\.[0-9]+\\.[0-9]+"),
			}),
		})

	// Verify - Run the actual test
	if err := pact.Verify(func() error {
		return nil // Same as above for testing
	}); err != nil {
		t.Fatal(err)
	}

	// Write pact files
	if err := pact.WritePact(); err != nil {
		t.Fatal(err)
	}
}

// Provider Verification (normally done in the provider's test suite)
func TestContentDeliveryProviderPact(t *testing.T) {
	// Skip this test if we're not running provider verification
	if os.Getenv("PACT_PROVIDER_VERIFICATION") != "true" {
		t.Skip("Skipping provider verification")
	}

	// Start the provider service (your actual service)
	// In a real scenario, you'd start your server

	// Verify the provider against previously created pacts
	pact := dsl.Pact{
		Provider: "content-delivery-service",
	}

	// Verify the provider using the Pact Verifier
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        "http://localhost:8082", // URL of the provider API
		PactURLs:               []string{fmt.Sprintf("%s/cms-service-content-delivery-service.json", pactURL)},
		ProviderStatesSetupURL: "http://localhost:8082/setup", // URL to setup provider states
	})

	if err != nil {
		t.Fatal(err)
	}
}
