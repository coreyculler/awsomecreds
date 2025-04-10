//go:build !skipintegration

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
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

	// Skip test if required environment variables are not set
	if roleArn == "" {
		t.Skip("Skipping integration test. TEST_ROLE_ARN environment variable must be set")
	}

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

// Test the generate command with integration
func TestGenerateCommandIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=1 to run")
	}

	// These values should be set in environment variables for CI/CD
	sourceProfile := os.Getenv("TEST_SOURCE_PROFILE")
	roleArn := os.Getenv("TEST_ROLE_ARN")
	mfaToken := os.Getenv("TEST_MFA_TOKEN") // Optional
	region := os.Getenv("TEST_REGION")      // Optional

	// Skip test if required environment variables are not set
	if roleArn == "" {
		t.Skip("Skipping integration test. TEST_ROLE_ARN environment variable must be set")
	}

	// Test case 1: Generate shell format output
	t.Run("ShellFormat", func(t *testing.T) {
		// Redirect stderr to avoid clutter in test output
		oldStderr := os.Stderr
		stderrR, stderrW, _ := os.Pipe()
		os.Stderr = stderrW

		// Capture stdout to check the output
		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		// Run the actual function
		err := outputTempCredentials(sourceProfile, roleArn, mfaToken, region, 3600, "shell")

		// Close the write end of the pipes to complete the capture
		stdoutW.Close()
		stderrW.Close()

		// Restore stdout and stderr
		os.Stdout = oldStdout
		os.Stderr = oldStderr

		// Read the captured output
		var stdoutBuf bytes.Buffer
		io.Copy(&stdoutBuf, stdoutR)

		// Discard stderr output
		io.Copy(io.Discard, stderrR)

		if err != nil {
			t.Errorf("Shell output generation failed: %v", err)
		}

		// Check the output contains the expected environment variables
		output := stdoutBuf.String()
		for _, expected := range []string{
			"export AWS_ACCESS_KEY_ID=",
			"export AWS_SECRET_ACCESS_KEY=",
			"export AWS_SESSION_TOKEN=",
			"export AWS_CREDENTIAL_EXPIRATION=",
		} {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain '%s', but it didn't.\nOutput: %s", expected, output)
			}
		}
	})

	// Test case 2: Generate JSON format output
	t.Run("JSONFormat", func(t *testing.T) {
		// Redirect stderr to avoid clutter in test output
		oldStderr := os.Stderr
		stderrR, stderrW, _ := os.Pipe()
		os.Stderr = stderrW

		// Capture stdout to check the output
		oldStdout := os.Stdout
		stdoutR, stdoutW, _ := os.Pipe()
		os.Stdout = stdoutW

		// Run the actual function
		err := outputTempCredentials(sourceProfile, roleArn, mfaToken, region, 3600, "json")

		// Close the write end of the pipes to complete the capture
		stdoutW.Close()
		stderrW.Close()

		// Restore stdout and stderr
		os.Stdout = oldStdout
		os.Stderr = oldStderr

		// Read the captured output
		var stdoutBuf bytes.Buffer
		io.Copy(&stdoutBuf, stdoutR)

		// Discard stderr output
		io.Copy(io.Discard, stderrR)

		if err != nil {
			t.Errorf("JSON output generation failed: %v", err)
		}

		// Check that the output is valid JSON and contains the expected fields
		output := stdoutBuf.String()
		var creds Credentials
		if err := json.Unmarshal([]byte(output), &creds); err != nil {
			t.Errorf("Failed to parse JSON output: %v\nOutput: %s", err, output)
		}

		// Verify credentials are not empty
		if creds.AccessKeyId == "" {
			t.Error("AccessKeyId is empty in JSON output")
		}
		if creds.SecretAccessKey == "" {
			t.Error("SecretAccessKey is empty in JSON output")
		}
		if creds.SessionToken == "" {
			t.Error("SessionToken is empty in JSON output")
		}
	})
}
