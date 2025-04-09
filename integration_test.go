//go:build !skipintegration

package main

import (
	"os"
	"testing"
)

// This test requires actual AWS credentials and will make real AWS API calls
// Run with: go test -tags=integration -v ./...
func TestIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=1 to run")
	}

	// These values should be set in environment variables for CI/CD
	sourceProfile := os.Getenv("TEST_SOURCE_PROFILE")
	roleArn := os.Getenv("TEST_ROLE_ARN")
	mfaToken := os.Getenv("TEST_MFA_TOKEN") // Optional
	newProfile := "awsomecreds-test-profile"

	// Run the actual function
	err := generateTempProfile(sourceProfile, roleArn, mfaToken, newProfile, "", 3600)
	if err != nil {
		t.Errorf("Integration test failed: %v", err)
	}

	// Clean up the test profile
	cmd := execCommand("aws", "configure", "rm", "--profile", newProfile, "aws_access_key_id")
	cmd.Run()
	cmd = execCommand("aws", "configure", "rm", "--profile", newProfile, "aws_secret_access_key")
	cmd.Run()
	cmd = execCommand("aws", "configure", "rm", "--profile", newProfile, "aws_session_token")
	cmd.Run()
	cmd = execCommand("aws", "configure", "rm", "--profile", newProfile, "region")
	cmd.Run()
}
